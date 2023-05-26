// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv64

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/ld"
)

func Init() (*sys.Arch, ld.Arch) {
	arch := sys.ArchRISCV64

	theArch := ld.Arch{
		Funcalign:  funcAlign,
		Maxalign:   maxAlign,
		Minalign:   minAlign,
		Dwarfregsp: dwarfRegSP,
		Dwarfreglr: dwarfRegLR,

		Archinit:         archinit,
		Archreloc:        archreloc,
		Archrelocvariant: archrelocvariant,
		Extreloc:         extreloc,
		Elfreloc1:        elfreloc1,
		ElfrelocSize:     24,
		Elfsetupplt:      elfsetupplt,

		// TrampLimit is set such that we always run the trampoline
		// generation code. This is necessary since calls to external
		// symbols require the use of trampolines, regardless of the
		// text size.
		TrampLimit: 1,
		Trampoline: trampoline,

		Gentext:     gentext,
		GenSymsLate: genSymsLate,
		Machoreloc1: machoreloc1,

		Linuxdynld: "/lib/ld.so.1",

		Freebsddynld:   "/usr/libexec/ld-elf.so.1",
		Netbsddynld:    "XXX",
		Openbsddynld:   "XXX",
		Dragonflydynld: "XXX",
		Solarisdynld:   "XXX",
	}

	return arch, theArch
}

func archinit(ctxt *ld.Link) {
	switch ctxt.HeadType {
	case objabi.Hlinux, objabi.Hfreebsd:
		ld.Elfinit(ctxt)
		ld.HEADR = ld.ELFRESERVE
		if *ld.FlagTextAddr == -1 {
			*ld.FlagTextAddr = 0x10000 + int64(ld.HEADR)
		}
		if *ld.FlagRound == -1 {
			*ld.FlagRound = 0x10000
		}
	default:
		ld.Exitf("unknown -H option: %v", ctxt.HeadType)
	}
}
