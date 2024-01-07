// Package gcsobj supplies a seekable Reader type for Google Cloud Storage objects.
package gcsobj

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"cloud.google.com/go/storage"
)

// Reader is an io.ReadSeeker for objects in Google Cloud Storage buckets.
type Reader struct {
	// Embedding a context in a data structure is an antipattern,
	// except when needed to satisfy interfaces (like io.ReadSeeker) that don't permit passing a context.
	// See https://go.dev/wiki/CodeReviewComments#contexts
	ctx context.Context

	obj       *storage.ObjectHandle
	r         *storage.Reader
	pos, size int64
	nread     int64 // Read/write with atomic
}

// NewReader creates a new Reader on the given object.
// If the object size is already known, use [NewReaderWithSize] instead.
// Callers must call the Close method when finished with the Reader.
func NewReader(ctx context.Context, obj *storage.ObjectHandle) (*Reader, error) {
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, err
	}
	return NewReaderWithSize(ctx, obj, attrs.Size), nil
}

// NewReaderWithSize creates a new Reader on the given object.
// Use this in preference to [NewReader] when the object size is already known
// (e.g. from an earlier call to [storage.ObjectHandle.Attrs]).
// Callers must call the Close method when finished with the Reader.
func NewReaderWithSize(ctx context.Context, obj *storage.ObjectHandle, size int64) *Reader {
	return &Reader{
		ctx:  ctx,
		obj:  obj,
		size: size,
	}
}

// Read implements io.Reader.
func (r *Reader) Read(dest []byte) (int, error) {
	if r.r == nil && r.pos < r.size {
		var err error
		r.r, err = r.obj.NewRangeReader(r.ctx, r.pos, -1)
		if err != nil {
			return 0, err
		}
	}
	if r.r == nil {
		return 0, io.EOF
	}
	n, err := r.r.Read(dest)
	r.pos += int64(n)
	atomic.AddInt64(&r.nread, int64(n))
	return n, err
}

// Seek implements io.Seeker.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	err := r.Close()
	if err != nil {
		return 0, err
	}

	switch whence {
	case io.SeekStart:
		r.pos = offset
	case io.SeekCurrent:
		r.pos += offset
	case io.SeekEnd:
		r.pos = r.size + offset
	default:
		return 0, fmt.Errorf("illegal whence value %d", whence)
	}

	return r.pos, nil
}

// Close closes a Reader and releases its resources.
func (r *Reader) Close() error {
	if r.r == nil {
		return nil
	}
	err := r.r.Close()
	r.r = nil
	return err
}

// NRead reports the number of bytes that have been read from Reader.
// This is safe to call concurrently with Read.
func (r *Reader) NRead() int64 {
	return atomic.LoadInt64(&r.nread)
}
