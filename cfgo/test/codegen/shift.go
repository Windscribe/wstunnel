// asmcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// ------------------ //
//   constant shifts  //
// ------------------ //

func lshConst64x64(v int64) int64 {
	// ppc64:"SLD"
	// ppc64le:"SLD"
	// riscv64:"SLLI",-"AND",-"SLTIU"
	return v << uint64(33)
}

func rshConst64Ux64(v uint64) uint64 {
	// ppc64:"SRD"
	// ppc64le:"SRD"
	// riscv64:"SRLI",-"AND",-"SLTIU"
	return v >> uint64(33)
}

func rshConst64x64(v int64) int64 {
	// ppc64:"SRAD"
	// ppc64le:"SRAD"
	// riscv64:"SRAI",-"OR",-"SLTIU"
	return v >> uint64(33)
}

func lshConst32x64(v int32) int32 {
	// ppc64:"SLW"
	// ppc64le:"SLW"
	// riscv64:"SLLI",-"AND",-"SLTIU", -"MOVW"
	return v << uint64(29)
}

func rshConst32Ux64(v uint32) uint32 {
	// ppc64:"SRW"
	// ppc64le:"SRW"
	// riscv64:"SRLI",-"AND",-"SLTIU", -"MOVW"
	return v >> uint64(29)
}

func rshConst32x64(v int32) int32 {
	// ppc64:"SRAW"
	// ppc64le:"SRAW"
	// riscv64:"SRAI",-"OR",-"SLTIU", -"MOVW"
	return v >> uint64(29)
}

func lshConst64x32(v int64) int64 {
	// ppc64:"SLD"
	// ppc64le:"SLD"
	// riscv64:"SLLI",-"AND",-"SLTIU"
	return v << uint32(33)
}

func rshConst64Ux32(v uint64) uint64 {
	// ppc64:"SRD"
	// ppc64le:"SRD"
	// riscv64:"SRLI",-"AND",-"SLTIU"
	return v >> uint32(33)
}

func rshConst64x32(v int64) int64 {
	// ppc64:"SRAD"
	// ppc64le:"SRAD"
	// riscv64:"SRAI",-"OR",-"SLTIU"
	return v >> uint32(33)
}

// ------------------ //
//   masked shifts    //
// ------------------ //

func lshMask64x64(v int64, s uint64) int64 {
	// arm64:"LSL",-"AND"
	// ppc64:"ANDCC",-"ORN",-"ISEL"
	// ppc64le:"ANDCC",-"ORN",-"ISEL"
	// riscv64:"SLL",-"AND\t",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v << (s & 63)
}

func rshMask64Ux64(v uint64, s uint64) uint64 {
	// arm64:"LSR",-"AND",-"CSEL"
	// ppc64:"ANDCC",-"ORN",-"ISEL"
	// ppc64le:"ANDCC",-"ORN",-"ISEL"
	// riscv64:"SRL",-"AND\t",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v >> (s & 63)
}

func rshMask64x64(v int64, s uint64) int64 {
	// arm64:"ASR",-"AND",-"CSEL"
	// ppc64:"ANDCC",-"ORN",-"ISEL"
	// ppc64le:"ANDCC",-ORN",-"ISEL"
	// riscv64:"SRA",-"OR",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v >> (s & 63)
}

func lshMask32x64(v int32, s uint64) int32 {
	// arm64:"LSL",-"AND"
	// ppc64:"ISEL",-"ORN"
	// ppc64le:"ISEL",-"ORN"
	// riscv64:"SLL",-"AND\t",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v << (s & 63)
}

func rshMask32Ux64(v uint32, s uint64) uint32 {
	// arm64:"LSR",-"AND"
	// ppc64:"ISEL",-"ORN"
	// ppc64le:"ISEL",-"ORN"
	// riscv64:"SRL",-"AND\t",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v >> (s & 63)
}

func rshMask32x64(v int32, s uint64) int32 {
	// arm64:"ASR",-"AND"
	// ppc64:"ISEL",-"ORN"
	// ppc64le:"ISEL",-"ORN"
	// riscv64:"SRA",-"OR",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v >> (s & 63)
}

func lshMask64x32(v int64, s uint32) int64 {
	// arm64:"LSL",-"AND"
	// ppc64:"ANDCC",-"ORN"
	// ppc64le:"ANDCC",-"ORN"
	// riscv64:"SLL",-"AND\t",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v << (s & 63)
}

func rshMask64Ux32(v uint64, s uint32) uint64 {
	// arm64:"LSR",-"AND",-"CSEL"
	// ppc64:"ANDCC",-"ORN"
	// ppc64le:"ANDCC",-"ORN"
	// riscv64:"SRL",-"AND\t",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v >> (s & 63)
}

func rshMask64x32(v int64, s uint32) int64 {
	// arm64:"ASR",-"AND",-"CSEL"
	// ppc64:"ANDCC",-"ORN",-"ISEL"
	// ppc64le:"ANDCC",-"ORN",-"ISEL"
	// riscv64:"SRA",-"OR",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v >> (s & 63)
}

func lshMask64x32Ext(v int64, s int32) int64 {
	// ppc64:"ANDCC",-"ORN",-"ISEL"
	// ppc64le:"ANDCC",-"ORN",-"ISEL"
	// riscv64:"SLL",-"AND\t",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v << uint(s&63)
}

func rshMask64Ux32Ext(v uint64, s int32) uint64 {
	// ppc64:"ANDCC",-"ORN",-"ISEL"
	// ppc64le:"ANDCC",-"ORN",-"ISEL"
	// riscv64:"SRL",-"AND\t",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v >> uint(s&63)
}

func rshMask64x32Ext(v int64, s int32) int64 {
	// ppc64:"ANDCC",-"ORN",-"ISEL"
	// ppc64le:"ANDCC",-"ORN",-"ISEL"
	// riscv64:"SRA",-"OR",-"SLTIU"
	// s390x:-"RISBGZ",-"AND",-"LOCGR"
	return v >> uint(s&63)
}

// --------------- //
//  signed shifts  //
// --------------- //

// We do want to generate a test + panicshift for these cases.
func lshSigned(v8 int8, v16 int16, v32 int32, v64 int64, x int) {
	// amd64:"TESTB"
	_ = x << v8
	// amd64:"TESTW"
	_ = x << v16
	// amd64:"TESTL"
	_ = x << v32
	// amd64:"TESTQ"
	_ = x << v64
}

// We want to avoid generating a test + panicshift for these cases.
func lshSignedMasked(v8 int8, v16 int16, v32 int32, v64 int64, x int) {
	// amd64:-"TESTB"
	_ = x << (v8 & 7)
	// amd64:-"TESTW"
	_ = x << (v16 & 15)
	// amd64:-"TESTL"
	_ = x << (v32 & 31)
	// amd64:-"TESTQ"
	_ = x << (v64 & 63)
}

// ------------------ //
//   bounded shifts   //
// ------------------ //

func lshGuarded64(v int64, s uint) int64 {
	if s < 64 {
		// riscv64:"SLL",-"AND",-"SLTIU"
		// s390x:-"RISBGZ",-"AND",-"LOCGR"
		// wasm:-"Select",-".*LtU"
		// arm64:"LSL",-"CSEL"
		return v << s
	}
	panic("shift too large")
}

func rshGuarded64U(v uint64, s uint) uint64 {
	if s < 64 {
		// riscv64:"SRL",-"AND",-"SLTIU"
		// s390x:-"RISBGZ",-"AND",-"LOCGR"
		// wasm:-"Select",-".*LtU"
		// arm64:"LSR",-"CSEL"
		return v >> s
	}
	panic("shift too large")
}

func rshGuarded64(v int64, s uint) int64 {
	if s < 64 {
		// riscv64:"SRA",-"OR",-"SLTIU"
		// s390x:-"RISBGZ",-"AND",-"LOCGR"
		// wasm:-"Select",-".*LtU"
		// arm64:"ASR",-"CSEL"
		return v >> s
	}
	panic("shift too large")
}

func provedUnsignedShiftLeft(val64 uint64, val32 uint32, val16 uint16, val8 uint8, shift int) (r1 uint64, r2 uint32, r3 uint16, r4 uint8) {
	if shift >= 0 && shift < 64 {
		// arm64:"LSL",-"CSEL"
		r1 = val64 << shift
	}
	if shift >= 0 && shift < 32 {
		// arm64:"LSL",-"CSEL"
		r2 = val32 << shift
	}
	if shift >= 0 && shift < 16 {
		// arm64:"LSL",-"CSEL"
		r3 = val16 << shift
	}
	if shift >= 0 && shift < 8 {
		// arm64:"LSL",-"CSEL"
		r4 = val8 << shift
	}
	return r1, r2, r3, r4
}

func provedSignedShiftLeft(val64 int64, val32 int32, val16 int16, val8 int8, shift int) (r1 int64, r2 int32, r3 int16, r4 int8) {
	if shift >= 0 && shift < 64 {
		// arm64:"LSL",-"CSEL"
		r1 = val64 << shift
	}
	if shift >= 0 && shift < 32 {
		// arm64:"LSL",-"CSEL"
		r2 = val32 << shift
	}
	if shift >= 0 && shift < 16 {
		// arm64:"LSL",-"CSEL"
		r3 = val16 << shift
	}
	if shift >= 0 && shift < 8 {
		// arm64:"LSL",-"CSEL"
		r4 = val8 << shift
	}
	return r1, r2, r3, r4
}

func provedUnsignedShiftRight(val64 uint64, val32 uint32, val16 uint16, val8 uint8, shift int) (r1 uint64, r2 uint32, r3 uint16, r4 uint8) {
	if shift >= 0 && shift < 64 {
		// arm64:"LSR",-"CSEL"
		r1 = val64 >> shift
	}
	if shift >= 0 && shift < 32 {
		// arm64:"LSR",-"CSEL"
		r2 = val32 >> shift
	}
	if shift >= 0 && shift < 16 {
		// arm64:"LSR",-"CSEL"
		r3 = val16 >> shift
	}
	if shift >= 0 && shift < 8 {
		// arm64:"LSR",-"CSEL"
		r4 = val8 >> shift
	}
	return r1, r2, r3, r4
}

func provedSignedShiftRight(val64 int64, val32 int32, val16 int16, val8 int8, shift int) (r1 int64, r2 int32, r3 int16, r4 int8) {
	if shift >= 0 && shift < 64 {
		// arm64:"ASR",-"CSEL"
		r1 = val64 >> shift
	}
	if shift >= 0 && shift < 32 {
		// arm64:"ASR",-"CSEL"
		r2 = val32 >> shift
	}
	if shift >= 0 && shift < 16 {
		// arm64:"ASR",-"CSEL"
		r3 = val16 >> shift
	}
	if shift >= 0 && shift < 8 {
		// arm64:"ASR",-"CSEL"
		r4 = val8 >> shift
	}
	return r1, r2, r3, r4
}

func checkUnneededTrunc(tab *[100000]uint32, d uint64, v uint32, h uint16, b byte) (uint32, uint64) {

	// ppc64le:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	// ppc64:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	f := tab[byte(v)^b]
	// ppc64le:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	// ppc64:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	f += tab[byte(v)&b]
	// ppc64le:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	// ppc64:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	f += tab[byte(v)|b]
	// ppc64le:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	// ppc64:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	f += tab[uint16(v)&h]
	// ppc64le:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	// ppc64:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	f += tab[uint16(v)^h]
	// ppc64le:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	// ppc64:-".*RLWINM",-".*RLDICR",".*CLRLSLDI"
	f += tab[uint16(v)|h]
	// ppc64le:-".*AND",-"RLDICR",".*CLRLSLDI"
	// ppc64:-".*AND",-"RLDICR",".*CLRLSLDI"
	f += tab[v&0xff]
	// ppc64le:-".*AND",".*CLRLSLWI"
	// ppc64:-".*AND",".*CLRLSLWI"
	f += 2 * uint32(uint16(d))
	// ppc64le:-".*AND",-"RLDICR",".*CLRLSLDI"
	// ppc64:-".*AND",-"RLDICR",".*CLRLSLDI"
	g := 2 * uint64(uint32(d))
	return f, g
}

func checkCombinedShifts(v8 uint8, v16 uint16, v32 uint32, x32 int32, v64 uint64) (uint8, uint16, uint32, uint64, int64) {

	// ppc64le:-"AND","CLRLSLWI"
	// ppc64:-"AND","CLRLSLWI"
	f := (v8 & 0xF) << 2
	// ppc64le:"CLRLSLWI"
	// ppc64:"CLRLSLWI"
	f += byte(v16) << 3
	// ppc64le:-"AND","CLRLSLWI"
	// ppc64:-"AND","CLRLSLWI"
	g := (v16 & 0xFF) << 3
	// ppc64le:-"AND","CLRLSLWI"
	// ppc64:-"AND","CLRLSLWI"
	h := (v32 & 0xFFFFF) << 2
	// ppc64le:"CLRLSLDI"
	// ppc64:"CLRLSLDI"
	i := (v64 & 0xFFFFFFFF) << 5
	// ppc64le:-"CLRLSLDI"
	// ppc64:-"CLRLSLDI"
	i += (v64 & 0xFFFFFFF) << 38
	// ppc64le/power9:-"CLRLSLDI"
	// ppc64/power9:-"CLRLSLDI"
	i += (v64 & 0xFFFF00) << 10
	// ppc64le/power9:-"SLD","EXTSWSLI"
	// ppc64/power9:-"SLD","EXTSWSLI"
	j := int64(x32+32) * 8
	return f, g, h, i, j
}

func checkWidenAfterShift(v int64, u uint64) (int64, uint64) {

	// ppc64le:-".*MOVW"
	f := int32(v >> 32)
	// ppc64le:".*MOVW"
	f += int32(v >> 31)
	// ppc64le:-".*MOVH"
	g := int16(v >> 48)
	// ppc64le:".*MOVH"
	g += int16(v >> 30)
	// ppc64le:-".*MOVH"
	g += int16(f >> 16)
	// ppc64le:-".*MOVB"
	h := int8(v >> 56)
	// ppc64le:".*MOVB"
	h += int8(v >> 28)
	// ppc64le:-".*MOVB"
	h += int8(f >> 24)
	// ppc64le:".*MOVB"
	h += int8(f >> 16)
	return int64(h), uint64(g)
}

func checkShiftAndMask32(v []uint32) {
	i := 0

	// ppc64le: "RLWNM\t[$]24, R[0-9]+, [$]12, [$]19, R[0-9]+"
	// ppc64: "RLWNM\t[$]24, R[0-9]+, [$]12, [$]19, R[0-9]+"
	v[i] = (v[i] & 0xFF00000) >> 8
	i++
	// ppc64le: "RLWNM\t[$]26, R[0-9]+, [$]22, [$]29, R[0-9]+"
	// ppc64: "RLWNM\t[$]26, R[0-9]+, [$]22, [$]29, R[0-9]+"
	v[i] = (v[i] & 0xFF00) >> 6
	i++
	// ppc64le: "MOVW\tR0"
	// ppc64: "MOVW\tR0"
	v[i] = (v[i] & 0xFF) >> 8
	i++
	// ppc64le: "MOVW\tR0"
	// ppc64: "MOVW\tR0"
	v[i] = (v[i] & 0xF000000) >> 28
	i++
	// ppc64le: "RLWNM\t[$]26, R[0-9]+, [$]24, [$]31, R[0-9]+"
	// ppc64: "RLWNM\t[$]26, R[0-9]+, [$]24, [$]31, R[0-9]+"
	v[i] = (v[i] >> 6) & 0xFF
	i++
	// ppc64le: "RLWNM\t[$]26, R[0-9]+, [$]12, [$]19, R[0-9]+"
	// ppc64: "RLWNM\t[$]26, R[0-9]+, [$]12, [$]19, R[0-9]+"
	v[i] = (v[i] >> 6) & 0xFF000
	i++
	// ppc64le: "MOVW\tR0"
	// ppc64: "MOVW\tR0"
	v[i] = (v[i] >> 20) & 0xFF000
	i++
	// ppc64le: "MOVW\tR0"
	// ppc64: "MOVW\tR0"
	v[i] = (v[i] >> 24) & 0xFF00
	i++
}

func checkMergedShifts32(a [256]uint32, b [256]uint64, u uint32, v uint32) {
	// ppc64le: -"CLRLSLDI", "RLWNM\t[$]10, R[0-9]+, [$]22, [$]29, R[0-9]+"
	// ppc64: -"CLRLSLDI", "RLWNM\t[$]10, R[0-9]+, [$]22, [$]29, R[0-9]+"
	a[0] = a[uint8(v>>24)]
	// ppc64le: -"CLRLSLDI", "RLWNM\t[$]11, R[0-9]+, [$]21, [$]28, R[0-9]+"
	// ppc64: -"CLRLSLDI", "RLWNM\t[$]11, R[0-9]+, [$]21, [$]28, R[0-9]+"
	b[0] = b[uint8(v>>24)]
	// ppc64le: -"CLRLSLDI", "RLWNM\t[$]15, R[0-9]+, [$]21, [$]28, R[0-9]+"
	// ppc64: -"CLRLSLDI", "RLWNM\t[$]15, R[0-9]+, [$]21, [$]28, R[0-9]+"
	b[1] = b[(v>>20)&0xFF]
	// ppc64le: -"SLD", "RLWNM\t[$]10, R[0-9]+, [$]22, [$]28, R[0-9]+"
	// ppc64: -"SLD", "RLWNM\t[$]10, R[0-9]+, [$]22, [$]28, R[0-9]+"
	b[2] = b[v>>25]
}

// 128 bit shifts

func check128bitShifts(x, y uint64, bits uint) (uint64, uint64) {
	s := bits & 63
	ŝ := (64 - bits) & 63
	// check that the shift operation has two commas (three operands)
	// amd64:"SHRQ.*,.*,"
	shr := x>>s | y<<ŝ
	// amd64:"SHLQ.*,.*,"
	shl := x<<s | y>>ŝ
	return shr, shl
}

func checkShiftToMask(u []uint64, s []int64) {
	// amd64:-"SHR",-"SHL","ANDQ"
	u[0] = u[0] >> 5 << 5
	// amd64:-"SAR",-"SHL","ANDQ"
	s[0] = s[0] >> 5 << 5
	// amd64:-"SHR",-"SHL","ANDQ"
	u[1] = u[1] << 5 >> 5
}
