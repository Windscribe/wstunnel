// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa_test

import (
	cmddwarf "cmd/internal/dwarf"
	"cmd/internal/quoted"
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"internal/testenv"
	"internal/xcoff"
	"io"
	"os"
	"runtime"
	"sort"
	"testing"
)

func open(path string) (*dwarf.Data, error) {
	if fh, err := elf.Open(path); err == nil {
		return fh.DWARF()
	}

	if fh, err := pe.Open(path); err == nil {
		return fh.DWARF()
	}

	if fh, err := macho.Open(path); err == nil {
		return fh.DWARF()
	}

	if fh, err := xcoff.Open(path); err == nil {
		return fh.DWARF()
	}

	return nil, fmt.Errorf("unrecognized executable format")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type Line struct {
	File string
	Line int
}

func TestStmtLines(t *testing.T) {
	if runtime.GOOS == "plan9" {
		t.Skip("skipping on plan9; no DWARF symbol table in executables")
	}

	if runtime.GOOS == "aix" {
		extld := os.Getenv("CC")
		if extld == "" {
			extld = "gcc"
		}
		extldArgs, err := quoted.Split(extld)
		if err != nil {
			t.Fatal(err)
		}
		enabled, err := cmddwarf.IsDWARFEnabledOnAIXLd(extldArgs)
		if err != nil {
			t.Fatal(err)
		}
		if !enabled {
			t.Skip("skipping on aix: no DWARF with ld version < 7.2.2 ")
		}
	}

	lines := map[Line]bool{}
	dw, err := open(testenv.GoToolPath(t))
	must(err)
	rdr := dw.Reader()
	rdr.Seek(0)
	for {
		e, err := rdr.Next()
		must(err)
		if e == nil {
			break
		}
		if e.Tag != dwarf.TagCompileUnit {
			continue
		}
		pkgname, _ := e.Val(dwarf.AttrName).(string)
		if pkgname == "runtime" {
			continue
		}
		if pkgname == "crypto/internal/nistec/fiat" {
			continue // golang.org/issue/49372
		}
		if e.Val(dwarf.AttrStmtList) == nil {
			continue
		}
		lrdr, err := dw.LineReader(e)
		must(err)

		var le dwarf.LineEntry

		for {
			err := lrdr.Next(&le)
			if err == io.EOF {
				break
			}
			must(err)
			fl := Line{le.File.Name, le.Line}
			lines[fl] = lines[fl] || le.IsStmt
		}
	}

	nonStmtLines := []Line{}
	for line, isstmt := range lines {
		if !isstmt {
			nonStmtLines = append(nonStmtLines, line)
		}
	}

	var m int
	if runtime.GOARCH == "amd64" {
		m = 1 // > 99% obtained on amd64, no backsliding
	} else if runtime.GOARCH == "riscv64" {
		m = 3 // XXX temporary update threshold to 97% for regabi
	} else {
		m = 2 // expect 98% elsewhere.
	}

	if len(nonStmtLines)*100 > m*len(lines) {
		t.Errorf("Saw too many (%s, > %d%%) lines without statement marks, total=%d, nostmt=%d ('-run TestStmtLines -v' lists failing lines)\n", runtime.GOARCH, m, len(lines), len(nonStmtLines))
	}
	t.Logf("Saw %d out of %d lines without statement marks", len(nonStmtLines), len(lines))
	if testing.Verbose() {
		sort.Slice(nonStmtLines, func(i, j int) bool {
			if nonStmtLines[i].File != nonStmtLines[j].File {
				return nonStmtLines[i].File < nonStmtLines[j].File
			}
			return nonStmtLines[i].Line < nonStmtLines[j].Line
		})
		for _, l := range nonStmtLines {
			t.Logf("%s:%d has no DWARF is_stmt mark\n", l.File, l.Line)
		}
	}
	t.Logf("total=%d, nostmt=%d\n", len(lines), len(nonStmtLines))
}
