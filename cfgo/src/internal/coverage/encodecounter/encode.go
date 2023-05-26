// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encodecounter

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"internal/coverage"
	"internal/coverage/slicewriter"
	"internal/coverage/stringtab"
	"internal/coverage/uleb128"
	"io"
	"os"
	"sort"
)

// This package contains APIs and helpers for encoding initial portions
// of the counter data files emitted at runtime when coverage instrumentation
// is enabled.  Counter data files may contain multiple segments; the file
// header and first segment are written via the "Write" method below, and
// additional segments can then be added using "AddSegment".

type CoverageDataWriter struct {
	stab    *stringtab.Writer
	w       *bufio.Writer
	tmp     []byte
	cflavor coverage.CounterFlavor
	segs    uint32
	debug   bool
}

func NewCoverageDataWriter(w io.Writer, flav coverage.CounterFlavor) *CoverageDataWriter {
	r := &CoverageDataWriter{
		stab: &stringtab.Writer{},
		w:    bufio.NewWriter(w),

		tmp:     make([]byte, 64),
		cflavor: flav,
	}
	r.stab.InitWriter()
	r.stab.Lookup("")
	return r
}

// CounterVisitor describes a helper object used during counter file
// writing; when writing counter data files, clients pass a
// CounterVisitor to the write/emit routines. The writers will then
// first invoke the visitor's NumFuncs() method to find out how many
// function's worth of data to write, then it will invoke VisitFuncs.
// The expectation is that the VisitFuncs method will then invoke the
// callback "f" with data for each function to emit to the file.
type CounterVisitor interface {
	NumFuncs() (int, error)
	VisitFuncs(f CounterVisitorFn) error
}

// CounterVisitorFn describes a callback function invoked when writing
// coverage counter data.
type CounterVisitorFn func(pkid uint32, funcid uint32, counters []uint32) error

// Write writes the contents of the count-data file to the writer
// previously supplied to NewCoverageDataWriter. Returns an error
// if something went wrong somewhere with the write.
func (cfw *CoverageDataWriter) Write(metaFileHash [16]byte, args map[string]string, visitor CounterVisitor) error {
	if err := cfw.writeHeader(metaFileHash); err != nil {
		return err
	}
	return cfw.AppendSegment(args, visitor)
}

func padToFourByteBoundary(ws *slicewriter.WriteSeeker) error {
	sz := len(ws.BytesWritten())
	zeros := []byte{0, 0, 0, 0}
	rem := uint32(sz) % 4
	if rem != 0 {
		pad := zeros[:(4 - rem)]
		if nw, err := ws.Write(pad); err != nil {
			return err
		} else if nw != len(pad) {
			return fmt.Errorf("error: short write")
		}
	}
	return nil
}

func (cfw *CoverageDataWriter) writeSegmentPreamble(args map[string]string, visitor CounterVisitor) error {
	var csh coverage.CounterSegmentHeader
	if nf, err := visitor.NumFuncs(); err != nil {
		return err
	} else {
		csh.FcnEntries = uint64(nf)
	}

	// Write string table and args to a byte slice (since we need
	// to capture offsets at various points), then emit the slice
	// once we are done.
	cfw.stab.Freeze()
	ws := &slicewriter.WriteSeeker{}
	if err := cfw.stab.Write(ws); err != nil {
		return err
	}
	csh.StrTabLen = uint32(len(ws.BytesWritten()))

	akeys := make([]string, 0, len(args))
	for k := range args {
		akeys = append(akeys, k)
	}
	sort.Strings(akeys)

	wrULEB128 := func(v uint) error {
		cfw.tmp = cfw.tmp[:0]
		cfw.tmp = uleb128.AppendUleb128(cfw.tmp, v)
		if _, err := ws.Write(cfw.tmp); err != nil {
			return err
		}
		return nil
	}

	// Count of arg pairs.
	if err := wrULEB128(uint(len(args))); err != nil {
		return err
	}
	// Arg pairs themselves.
	for _, k := range akeys {
		ki := uint(cfw.stab.Lookup(k))
		if err := wrULEB128(ki); err != nil {
			return err
		}
		v := args[k]
		vi := uint(cfw.stab.Lookup(v))
		if err := wrULEB128(vi); err != nil {
			return err
		}
	}
	if err := padToFourByteBoundary(ws); err != nil {
		return err
	}
	csh.ArgsLen = uint32(len(ws.BytesWritten())) - csh.StrTabLen

	if cfw.debug {
		fmt.Fprintf(os.Stderr, "=-= counter segment header: %+v", csh)
		fmt.Fprintf(os.Stderr, " FcnEntries=0x%x StrTabLen=0x%x ArgsLen=0x%x\n",
			csh.FcnEntries, csh.StrTabLen, csh.ArgsLen)
	}

	// At this point we can now do the actual write.
	if err := binary.Write(cfw.w, binary.LittleEndian, csh); err != nil {
		return err
	}
	if err := cfw.writeBytes(ws.BytesWritten()); err != nil {
		return err
	}
	return nil
}

// AppendSegment appends a new segment to a counter data, with a new
// args section followed by a payload of counter data clauses.
func (cfw *CoverageDataWriter) AppendSegment(args map[string]string, visitor CounterVisitor) error {
	cfw.stab = &stringtab.Writer{}
	cfw.stab.InitWriter()
	cfw.stab.Lookup("")

	var err error
	for k, v := range args {
		cfw.stab.Lookup(k)
		cfw.stab.Lookup(v)
	}

	if err = cfw.writeSegmentPreamble(args, visitor); err != nil {
		return err
	}
	if err = cfw.writeCounters(visitor); err != nil {
		return err
	}
	if err = cfw.writeFooter(); err != nil {
		return err
	}
	if err := cfw.w.Flush(); err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	cfw.stab = nil
	return nil
}

func (cfw *CoverageDataWriter) writeHeader(metaFileHash [16]byte) error {
	// Emit file header.
	ch := coverage.CounterFileHeader{
		Magic:     coverage.CovCounterMagic,
		Version:   coverage.CounterFileVersion,
		MetaHash:  metaFileHash,
		CFlavor:   cfw.cflavor,
		BigEndian: false,
	}
	if err := binary.Write(cfw.w, binary.LittleEndian, ch); err != nil {
		return err
	}
	return nil
}

func (cfw *CoverageDataWriter) writeBytes(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	nw, err := cfw.w.Write(b)
	if err != nil {
		return fmt.Errorf("error writing counter data: %v", err)
	}
	if len(b) != nw {
		return fmt.Errorf("error writing counter data: short write")
	}
	return nil
}

func (cfw *CoverageDataWriter) writeCounters(visitor CounterVisitor) error {
	// Notes:
	// - this version writes everything little-endian, which means
	//   a call is needed to encode every value (expensive)
	// - we may want to move to a model in which we just blast out
	//   all counters, or possibly mmap the file and do the write
	//   implicitly.
	ctrb := make([]byte, 4)
	wrval := func(val uint32) error {
		var buf []byte
		var towr int
		if cfw.cflavor == coverage.CtrRaw {
			binary.LittleEndian.PutUint32(ctrb, val)
			buf = ctrb
			towr = 4
		} else if cfw.cflavor == coverage.CtrULeb128 {
			cfw.tmp = cfw.tmp[:0]
			cfw.tmp = uleb128.AppendUleb128(cfw.tmp, uint(val))
			buf = cfw.tmp
			towr = len(buf)
		} else {
			panic("internal error: bad counter flavor")
		}
		if sz, err := cfw.w.Write(buf); err != nil {
			return err
		} else if sz != towr {
			return fmt.Errorf("writing counters: short write")
		}
		return nil
	}

	// Write out entries for each live function.
	emitter := func(pkid uint32, funcid uint32, counters []uint32) error {
		if err := wrval(uint32(len(counters))); err != nil {
			return err
		}

		if err := wrval(pkid); err != nil {
			return err
		}

		if err := wrval(funcid); err != nil {
			return err
		}
		for _, val := range counters {
			if err := wrval(val); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visitor.VisitFuncs(emitter); err != nil {
		return err
	}
	return nil
}

func (cfw *CoverageDataWriter) writeFooter() error {
	cfw.segs++
	cf := coverage.CounterFileFooter{
		Magic:       coverage.CovCounterMagic,
		NumSegments: cfw.segs,
	}
	if err := binary.Write(cfw.w, binary.LittleEndian, cf); err != nil {
		return err
	}
	return nil
}
