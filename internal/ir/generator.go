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

// IteratorNext is a hand-driven `x.next()` through an IterableIterator<T> parameter a generator
// feeds (#qk8rztp S2c-2b): the result object iterator.next() gives, owned by the caller, as Node's
// is, a fresh { value, done } each call. Iterator is the parameter's read, through the narrowing's
// Defined or not. A generator steps, and the value it yields, held as Of (a number, a boolean or a
// reference), is moved into the result; once its body has finished, or a throw has left it, the
// result is done. A collection's iterator, when Collection says one feeds the parameter too, runs its
// own next. Function is the one generator function whose frames reach it, whose ResumeMayThrow says
// whether the caller tests for a throw (IteratorNextMayThrow); natively a throw builds no result.
// Site is where the next is, which the loud default for anything else names. Lowering makes it only
// by rewriting the S2c-2a CallClosure over next once the origins are proven (lower/iterable_views.go).
type IteratorNext struct {
	Iterator   Expression
	Function   int
	Of         Type
	Collection bool
	Site       string
}

func (IteratorNext) Type() Type { return Object }

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

// IsCollectionIterator is whether the object an iterable view holds is a Map's or a Set's library
// iterator, read from the runtime object itself, never its static type: the view's dispatch steps
// one only by the library iterator's own step (lower/library_map_set.go's libraryIteratorStep).
type IsCollectionIterator struct{ Value Expression }

func (IsCollectionIterator) Type() Type { return Boolean }

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

// IteratorNextMayThrow is a hand-driven next() that can leave a throw pending: one whose generator's
// body a throw can leave. A collection's iterator's next throws nothing.
func (p *Program) IteratorNextMayThrow(next IteratorNext) bool {
	return p.Functions[next.Function].ResumeMayThrow
}
