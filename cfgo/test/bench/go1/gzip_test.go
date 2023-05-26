// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This benchmark tests gzip and gunzip performance.

package go1

import (
	"bytes"
	gz "compress/gzip"
	"io"
	"testing"
)

func makeGunzip(jsonbytes []byte) []byte {
	return bytes.Repeat(jsonbytes, 10)
}

func makeGzip(jsongunz []byte) []byte {
	var buf bytes.Buffer
	c := gz.NewWriter(&buf)
	c.Write(jsongunz)
	c.Close()
	return buf.Bytes()
}

func gzip(jsongunz []byte) {
	c := gz.NewWriter(io.Discard)
	if _, err := c.Write(jsongunz); err != nil {
		panic(err)
	}
	if err := c.Close(); err != nil {
		panic(err)
	}
}

func gunzip(jsongz []byte) {
	r, err := gz.NewReader(bytes.NewBuffer(jsongz))
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(io.Discard, r); err != nil {
		panic(err)
	}
	r.Close()
}

func BenchmarkGzip(b *testing.B) {
	jsonbytes := makeJsonBytes()
	jsongunz := makeGunzip(jsonbytes)
	b.ResetTimer()
	b.SetBytes(int64(len(jsongunz)))
	for i := 0; i < b.N; i++ {
		gzip(jsongunz)
	}
}

func BenchmarkGunzip(b *testing.B) {
	jsonbytes := makeJsonBytes()
	jsongunz := makeGunzip(jsonbytes)
	jsongz := makeGzip(jsongunz)
	b.ResetTimer()
	b.SetBytes(int64(len(jsongunz)))
	for i := 0; i < b.N; i++ {
		gunzip(jsongz)
	}
}
