// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Debug arguments, set by -d flag.

package base

// Debug holds the parsed debugging configuration values.
var Debug DebugFlags

// DebugFlags defines the debugging configuration values (see var Debug).
// Each struct field is a different value, named for the lower-case of the field name.
// Each field must be an int or string and must have a `help` struct tag.
//
// The -d option takes a comma-separated list of settings.
// Each setting is name=value; for ints, name is short for name=1.
type DebugFlags struct {
	Append                int    `help:"print information about append compilation"`
	Checkptr              int    `help:"instrument unsafe pointer conversions\n0: instrumentation disabled\n1: conversions involving unsafe.Pointer are instrumented\n2: conversions to unsafe.Pointer force heap allocation" concurrent:"ok"`
	Closure               int    `help:"print information about closure compilation"`
	DclStack              int    `help:"run internal dclstack check"`
	Defer                 int    `help:"print information about defer compilation"`
	DisableNil            int    `help:"disable nil checks" concurrent:"ok"`
	DumpPtrs              int    `help:"show Node pointers values in dump output"`
	DwarfInl              int    `help:"print information about DWARF inlined function creation"`
	Export                int    `help:"print export data"`
	Fmahash               string `help:"hash value for use in debugging platform-dependent multiply-add use" concurrent:"ok"`
	GCAdjust              int    `help:"log adjustments to GOGC" concurrent:"ok"`
	GCCheck               int    `help:"check heap/gc use by compiler" concurrent:"ok"`
	GCProg                int    `help:"print dump of GC programs"`
	Gossahash             string `help:"hash value for use in debugging the compiler"`
	InlFuncsWithClosures  int    `help:"allow functions with closures to be inlined" concurrent:"ok"`
	InlStaticInit         int    `help:"allow static initialization of inlined calls" concurrent:"ok"`
	InterfaceCycles       int    `help:"allow anonymous interface cycles"`
	Libfuzzer             int    `help:"enable coverage instrumentation for libfuzzer"`
	LocationLists         int    `help:"print information about DWARF location list creation"`
	Nil                   int    `help:"print information about nil checks"`
	NoOpenDefer           int    `help:"disable open-coded defers" concurrent:"ok"`
	NoRefName             int    `help:"do not include referenced symbol names in object file" concurrent:"ok"`
	PCTab                 string `help:"print named pc-value table\nOne of: pctospadj, pctofile, pctoline, pctoinline, pctopcdata"`
	Panic                 int    `help:"show all compiler panics"`
	Reshape               int    `help:"print information about expression reshaping"`
	Shapify               int    `help:"print information about shaping recursive types"`
	Slice                 int    `help:"print information about slice compilation"`
	SoftFloat             int    `help:"force compiler to emit soft-float code" concurrent:"ok"`
	SyncFrames            int    `help:"how many writer stack frames to include at sync points in unified export data"`
	TypeAssert            int    `help:"print information about type assertion inlining"`
	TypecheckInl          int    `help:"eager typechecking of inline function bodies" concurrent:"ok"`
	Unified               int    `help:"enable unified IR construction"`
	WB                    int    `help:"print information about write barriers"`
	ABIWrap               int    `help:"print information about ABI wrapper generation"`
	MayMoreStack          string `help:"call named function before all stack growth checks" concurrent:"ok"`
	PGOInlineCDFThreshold string `help:"cummulative threshold percentage for determining call sites as hot candidates for inlining" concurrent:"ok"`
	PGOInlineBudget       int    `help:"inline budget for hot functions" concurrent:"ok"`
	PGOInline             int    `help:"debug profile-guided inlining"`

	ConcurrentOk bool // true if only concurrentOk flags seen
}

// DebugSSA is called to set a -d ssa/... option.
// If nil, those options are reported as invalid options.
// If DebugSSA returns a non-empty string, that text is reported as a compiler error.
var DebugSSA func(phase, flag string, val int, valString string) string
