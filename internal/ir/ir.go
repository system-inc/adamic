// Package ir is Adamic's intermediate representation: what the backends compile, with every type
// already proven by the checker.
//
// It holds exactly what the programs stage 0 can compile need, and grows one program at a time
// (docs/0.1.md). Nothing in it is speculative: an instruction exists because a fixture lowers to it
// and the oracle has checked what it prints.
package ir

// Program is one compiled Adamic program.
type Program struct {
	// Source is the entry file's base name, as written, for the header of what the backends emit.
	Source string

	// Strings are the program's string constants, as UTF-8, in first-use order.
	Strings []string

	// Main is what the program does, in order.
	Main []Instruction
}

// Instruction is one step of a program.
type Instruction interface {
	instruction()
}

// Stream is where a line is written.
type Stream int

const (
	Stdout Stream = 1
	Stderr Stream = 2
)

// WriteLine writes a string constant and a newline: console.log and console.error with one string.
type WriteLine struct {
	Stream Stream
	String int
}

func (WriteLine) instruction() {}
