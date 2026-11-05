package ir

import "strings"

// RegExpProgram contains constants compiled and checked before native emission.
type RegExpProgram struct{ Pattern, Flags, Declarations string }
type RegExpNew struct {
	Index, Source, Flags int
	Arguments            []Expression
}

func (RegExpNew) Type() Type { return Object }

// RegExpCall covers both RegExp and String methods. Value is evaluated first.
type RegExpCall struct {
	Value       Expression
	Arguments   []Expression
	Method      string
	Returns     Type
	Replacement *RegExpReplacement
}

func (r RegExpCall) Type() Type { return r.Returns }

// GlobalFlagMayThrow says a matchAll or replaceAll can throw Node's TypeError for a RegExp without
// the g flag (runtime/regexp.c and regexp_replace.c set it on the pending word): unless the pattern
// is a literal made right there with it, whose flags can't change. Lowering counts the calls that
// can, the flow graph gives them an exception edge, and the emitter tests the word after each one.
func (p *Program) GlobalFlagMayThrow(call RegExpCall) bool {
	// Given no argument there is no RegExp to lack the flag: matchAll() and replaceAll() given nothing
	// don't throw this, and an IR-built call may carry none (fresh's regexp_test.go).
	if (call.Method != "matchAll" && call.Method != "replaceAll") || len(call.Arguments) == 0 {
		return false
	}
	made, literal := call.Arguments[0].(RegExpNew)
	return !literal || !strings.Contains(p.Regexps[made.Index].Flags, "g")
}

// RegExpProperty reads result-array metadata, preserving optional access.
type RegExpProperty struct {
	Array    Expression
	Name     string
	Of       Type
	Optional bool
}

func (r RegExpProperty) Type() Type {
	if r.Optional {
		return Maybe(r.Of)
	}
	return r.Of
}

// Null is distinct in JavaScript; nullable references use a null pointer in C.
type Null struct{ Of Type }

func (n Null) Type() Type {
	if n.Of != 0 {
		return n.Of
	}
	return Object
}

type IsNull struct {
	Value       Expression
	AlwaysFalse bool
}

func (IsNull) Type() Type { return Boolean }

// RegExpGroup reads a declared named-group dictionary; absent keys are undefined.
type RegExpGroup struct {
	Object   Expression
	Name     string
	Of       Type
	Optional bool
}

func (r RegExpGroup) Type() Type { return r.Of }

// RegExpReplacement describes the closure ABI; Rest is its final array element type.
type RegExpReplacement struct {
	Groups []RegExpReplacementGroup

	Parameters []Type
	Kinds      []int
	RestKind   int
	Rest       Type
	Returns    Type
}

type RegExpReplacementGroup struct {
	Argument int
	Rest     bool
	Name     string
	Optional bool
}
