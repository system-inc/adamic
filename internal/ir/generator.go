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
type GeneratorNext struct {
	Generator Expression
	Function  int
	Of        Type
}

func (g GeneratorNext) Type() Type { return g.Of }

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
