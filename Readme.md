# Gcsobj - Seekable readers for Google Cloud Storage objects

[![Go Reference](https://pkg.go.dev/badge/github.com/bobg/gcsobj.svg)](https://pkg.go.dev/github.com/bobg/gcsobj)
[![Go Report Card](https://goreportcard.com/badge/github.com/bobg/gcsobj)](https://goreportcard.com/report/github.com/bobg/gcsobj)

This is gcsobj,
a Go package that wraps the [Reader](https://pkg.go.dev/cloud.google.com/go/storage#Reader) type from `cloud.google.com/go/storage`,
which implements only the `io.Reader` interface,
to add a `Seek` method,
satisfying the `io.ReadSeeker` interface.

Among other things,
this makes Google Cloud Storage objects suitable for use
with the standard Go [http.ServeContent](https://pkg.go.dev/net/http#ServeContent) function.

## Usage

```go
bucket := gcsClient.Bucket(bucketName)
obj := bucket.Object(objName)
reader, err := gcsobj.NewReader(ctx, obj)
if err != nil { ... }
defer reader.Close()

// ...use reader as an io.ReadSeeker...
```
