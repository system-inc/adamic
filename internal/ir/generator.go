package ir

// Yield hands its value to the generator's consumer and suspends the body until the next
// GeneratorNext. It appears only as Evaluate{Yield}, a statement: a yield used as a value is not yet
// built, since what next(x) sends in is not modeled (#qk8rztp S5).
type Yield struct {
	Site  int // one-based source site, resolved by lowering only for diagnostics
	Value Expression
}

func (Yield) Type() Type { return 0 }

// GeneratorNext runs the generator Generator holds to its next yield and is the value yielded,
// owned. Once the body has finished, or a throw has left it, it is the zero value of Of and
// GeneratorDone says so; a throw leaves the pending exception set, which the caller tests first when
// Function's ResumeMayThrow says one can. Function is the generator function the frame was made by.
// Generator is a Generator, or, in an iterable view's dispatch only, the Object-typed local the view
// holds, after IsGenerator said it is one (checkGeneratorRoads).
type GeneratorNext struct {
	Generator Expression
	Function  int
	Of        Type
}

func (g GeneratorNext) Type() Type { return g.Of }

// IterableOf is a generator held where a slot typed Iterable<T> or IterableIterator<T> is, an
// object: the only door from a Generator value to an Object-typed slot (Compiler, on #qk8rztp S2c-1).
// It is a direct call's argument and nothing else (checkGeneratorRoads), and it adds no retain or
// release of its own: the call's result is the caller's temporary, the callee borrows it, and the
// caller lets go of it after the call. Natively it is a pointer cast; on JavaScript, the value.
type IterableOf struct {
	Value Expression
	Site  int // one-based site, resolved by lowering only for diagnostics
}

func (IterableOf) Type() Type { return Object }

// IsGenerator is whether the object an iterable view holds is a generator, read from its heap kind,
// never its static type: the view's per-step dispatch (lower/iterable_views.go) steps one only by
// GeneratorNext and GeneratorDone.
type IsGenerator struct{ Value Expression }

func (IsGenerator) Type() Type { return Boolean }

// GeneratorDone is whether the generator's last GeneratorNext found the body finished.
type GeneratorDone struct {
	Generator Expression
}

func (GeneratorDone) Type() Type { return Boolean }

// HasGenerators is a program with a generator function.
func (p *Program) HasGenerators() bool {
	for _, function := range p.Functions {
		if function.Generator {
			return true
		}
	}
	return false
}

// GeneratorNextMayThrow is a step that can leave a throw pending: one whose body a throw can leave.
func (p *Program) GeneratorNextMayThrow(next GeneratorNext) bool {
	return p.Functions[next.Function].ResumeMayThrow
}
