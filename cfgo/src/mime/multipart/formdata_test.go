// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multipart

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/textproto"
	"os"
	"strings"
	"testing"
)

func TestReadForm(t *testing.T) {
	b := strings.NewReader(strings.ReplaceAll(message, "\n", "\r\n"))
	r := NewReader(b, boundary)
	f, err := r.ReadForm(25)
	if err != nil {
		t.Fatal("ReadForm:", err)
	}
	defer f.RemoveAll()
	if g, e := f.Value["texta"][0], textaValue; g != e {
		t.Errorf("texta value = %q, want %q", g, e)
	}
	if g, e := f.Value["textb"][0], textbValue; g != e {
		t.Errorf("texta value = %q, want %q", g, e)
	}
	fd := testFile(t, f.File["filea"][0], "filea.txt", fileaContents)
	if _, ok := fd.(*os.File); ok {
		t.Error("file is *os.File, should not be")
	}
	fd.Close()
	fd = testFile(t, f.File["fileb"][0], "fileb.txt", filebContents)
	if _, ok := fd.(*os.File); !ok {
		t.Errorf("file has unexpected underlying type %T", fd)
	}
	fd.Close()
}

func TestReadFormWithNamelessFile(t *testing.T) {
	b := strings.NewReader(strings.ReplaceAll(messageWithFileWithoutName, "\n", "\r\n"))
	r := NewReader(b, boundary)
	f, err := r.ReadForm(25)
	if err != nil {
		t.Fatal("ReadForm:", err)
	}
	defer f.RemoveAll()

	if g, e := f.Value["hiddenfile"][0], filebContents; g != e {
		t.Errorf("hiddenfile value = %q, want %q", g, e)
	}
}

// Issue 40430: Handle ReadForm(math.MaxInt64)
func TestReadFormMaxMemoryOverflow(t *testing.T) {
	b := strings.NewReader(strings.ReplaceAll(messageWithTextContentType, "\n", "\r\n"))
	r := NewReader(b, boundary)
	f, err := r.ReadForm(math.MaxInt64)
	if err != nil {
		t.Fatalf("ReadForm(MaxInt64): %v", err)
	}
	if f == nil {
		t.Fatal("ReadForm(MaxInt64): missing form")
	}
}

func TestReadFormWithTextContentType(t *testing.T) {
	// From https://github.com/golang/go/issues/24041
	b := strings.NewReader(strings.ReplaceAll(messageWithTextContentType, "\n", "\r\n"))
	r := NewReader(b, boundary)
	f, err := r.ReadForm(25)
	if err != nil {
		t.Fatal("ReadForm:", err)
	}
	defer f.RemoveAll()

	if g, e := f.Value["texta"][0], textaValue; g != e {
		t.Errorf("texta value = %q, want %q", g, e)
	}
}

func testFile(t *testing.T, fh *FileHeader, efn, econtent string) File {
	if fh.Filename != efn {
		t.Errorf("filename = %q, want %q", fh.Filename, efn)
	}
	if fh.Size != int64(len(econtent)) {
		t.Errorf("size = %d, want %d", fh.Size, len(econtent))
	}
	f, err := fh.Open()
	if err != nil {
		t.Fatal("opening file:", err)
	}
	b := new(strings.Builder)
	_, err = io.Copy(b, f)
	if err != nil {
		t.Fatal("copying contents:", err)
	}
	if g := b.String(); g != econtent {
		t.Errorf("contents = %q, want %q", g, econtent)
	}
	return f
}

const (
	fileaContents = "This is a test file."
	filebContents = "Another test file."
	textaValue    = "foo"
	textbValue    = "bar"
	boundary      = `MyBoundary`
)

const messageWithFileWithoutName = `
--MyBoundary
Content-Disposition: form-data; name="hiddenfile"; filename=""
Content-Type: text/plain

` + filebContents + `
--MyBoundary--
`

const messageWithTextContentType = `
--MyBoundary
Content-Disposition: form-data; name="texta"
Content-Type: text/plain

` + textaValue + `
--MyBoundary
`

const message = `
--MyBoundary
Content-Disposition: form-data; name="filea"; filename="filea.txt"
Content-Type: text/plain

` + fileaContents + `
--MyBoundary
Content-Disposition: form-data; name="fileb"; filename="fileb.txt"
Content-Type: text/plain

` + filebContents + `
--MyBoundary
Content-Disposition: form-data; name="texta"

` + textaValue + `
--MyBoundary
Content-Disposition: form-data; name="textb"

` + textbValue + `
--MyBoundary--
`

func TestReadForm_NoReadAfterEOF(t *testing.T) {
	maxMemory := int64(32) << 20
	boundary := `---------------------------8d345eef0d38dc9`
	body := `
-----------------------------8d345eef0d38dc9
Content-Disposition: form-data; name="version"

171
-----------------------------8d345eef0d38dc9--`

	mr := NewReader(&failOnReadAfterErrorReader{t: t, r: strings.NewReader(body)}, boundary)

	f, err := mr.ReadForm(maxMemory)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Got: %#v", f)
}

// failOnReadAfterErrorReader is an io.Reader wrapping r.
// It fails t if any Read is called after a failing Read.
type failOnReadAfterErrorReader struct {
	t      *testing.T
	r      io.Reader
	sawErr error
}

func (r *failOnReadAfterErrorReader) Read(p []byte) (n int, err error) {
	if r.sawErr != nil {
		r.t.Fatalf("unexpected Read on Reader after previous read saw error %v", r.sawErr)
	}
	n, err = r.r.Read(p)
	r.sawErr = err
	return
}

// TestReadForm_NonFileMaxMemory asserts that the ReadForm maxMemory limit is applied
// while processing non-file form data as well as file form data.
func TestReadForm_NonFileMaxMemory(t *testing.T) {
	n := 10<<20 + 25
	if testing.Short() {
		n = 10<<10 + 25
	}
	largeTextValue := strings.Repeat("1", n)
	message := `--MyBoundary
Content-Disposition: form-data; name="largetext"

` + largeTextValue + `
--MyBoundary--
`

	testBody := strings.ReplaceAll(message, "\n", "\r\n")
	testCases := []struct {
		name      string
		maxMemory int64
		err       error
	}{
		{"smaller", 50 + int64(len("largetext")) + 100, nil},
		{"exact-fit", 25 + int64(len("largetext")) + 100, nil},
		{"too-large", 0, ErrMessageTooLarge},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.maxMemory == 0 && testing.Short() {
				t.Skip("skipping in -short mode")
			}
			b := strings.NewReader(testBody)
			r := NewReader(b, boundary)
			f, err := r.ReadForm(tc.maxMemory)
			if err == nil {
				defer f.RemoveAll()
			}
			if tc.err != err {
				t.Fatalf("ReadForm error - got: %v; expected: %v", err, tc.err)
			}
			if err == nil {
				if g := f.Value["largetext"][0]; g != largeTextValue {
					t.Errorf("largetext mismatch: got size: %v, expected size: %v", len(g), len(largeTextValue))
				}
			}
		})
	}
}

// TestReadForm_MetadataTooLarge verifies that we account for the size of field names,
// MIME headers, and map entry overhead while limiting the memory consumption of parsed forms.
func TestReadForm_MetadataTooLarge(t *testing.T) {
	for _, test := range []struct {
		name string
		f    func(*Writer)
	}{{
		name: "large name",
		f: func(fw *Writer) {
			name := strings.Repeat("a", 10<<20)
			w, _ := fw.CreateFormField(name)
			w.Write([]byte("value"))
		},
	}, {
		name: "large MIME header",
		f: func(fw *Writer) {
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", `form-data; name="a"`)
			h.Set("X-Foo", strings.Repeat("a", 10<<20))
			w, _ := fw.CreatePart(h)
			w.Write([]byte("value"))
		},
	}, {
		name: "many parts",
		f: func(fw *Writer) {
			for i := 0; i < 110000; i++ {
				w, _ := fw.CreateFormField("f")
				w.Write([]byte("v"))
			}
		},
	}} {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			fw := NewWriter(&buf)
			test.f(fw)
			if err := fw.Close(); err != nil {
				t.Fatal(err)
			}
			fr := NewReader(&buf, fw.Boundary())
			_, err := fr.ReadForm(0)
			if err != ErrMessageTooLarge {
				t.Errorf("fr.ReadForm() = %v, want ErrMessageTooLarge", err)
			}
		})
	}
}

// TestReadForm_ManyFiles_Combined tests that a multipart form containing many files only
// results in a single on-disk file.
func TestReadForm_ManyFiles_Combined(t *testing.T) {
	const distinct = false
	testReadFormManyFiles(t, distinct)
}

// TestReadForm_ManyFiles_Distinct tests that setting GODEBUG=multipartfiles=distinct
// results in every file in a multipart form being placed in a distinct on-disk file.
func TestReadForm_ManyFiles_Distinct(t *testing.T) {
	t.Setenv("GODEBUG", "multipartfiles=distinct")
	const distinct = true
	testReadFormManyFiles(t, distinct)
}

func testReadFormManyFiles(t *testing.T, distinct bool) {
	var buf bytes.Buffer
	fw := NewWriter(&buf)
	const numFiles = 10
	for i := 0; i < numFiles; i++ {
		name := fmt.Sprint(i)
		w, err := fw.CreateFormFile(name, name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(name))
	}
	if err := fw.Close(); err != nil {
		t.Fatal(err)
	}
	fr := NewReader(&buf, fw.Boundary())
	fr.tempDir = t.TempDir()
	form, err := fr.ReadForm(0)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < numFiles; i++ {
		name := fmt.Sprint(i)
		if got := len(form.File[name]); got != 1 {
			t.Fatalf("form.File[%q] has %v entries, want 1", name, got)
		}
		fh := form.File[name][0]
		file, err := fh.Open()
		if err != nil {
			t.Fatalf("form.File[%q].Open() = %v", name, err)
		}
		if distinct {
			if _, ok := file.(*os.File); !ok {
				t.Fatalf("form.File[%q].Open: %T, want *os.File", name, file)
			}
		}
		got, err := io.ReadAll(file)
		file.Close()
		if string(got) != name || err != nil {
			t.Fatalf("read form.File[%q]: %q, %v; want %q, nil", name, string(got), err, name)
		}
	}
	dir, err := os.Open(fr.tempDir)
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()
	names, err := dir.Readdirnames(0)
	if err != nil {
		t.Fatal(err)
	}
	wantNames := 1
	if distinct {
		wantNames = numFiles
	}
	if len(names) != wantNames {
		t.Fatalf("temp dir contains %v files; want 1", len(names))
	}
	if err := form.RemoveAll(); err != nil {
		t.Fatalf("form.RemoveAll() = %v", err)
	}
	names, err = dir.Readdirnames(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("temp dir contains %v files; want 0", len(names))
	}
}
