// Package fuzz generates random valid Adamic 0.1 programs, runs each the three ways the oracle does
// (its source on Node, native under ASan and UBSan, the JavaScript backend on Node), and shrinks any
// disagreement or sanitizer finding to a minimal program by delta debugging.
//
// Programs are trees rather than text, so the shrinker can remove a statement or replace an
// expression with a simpler one of the same type and print the result again.
package fuzz

import "strings"

// Type is what an expression evaluates to, as far as the generator needs to know: enough to put a
// value only where the checker will accept it, and to swap one expression for another in shrinking.
type Type string

const (
	Number       Type = "number"
	String       Type = "string"
	Boolean      Type = "boolean"
	NumberArray  Type = "number[]"
	StringArray  Type = "string[]"
	NumberMap    Type = "Map<string, number>"
	NumberKeyMap Type = "Map<number, number>"
	NumberSet    Type = "Set<number>"
	// Other is any type the shrinker never substitutes: a closure, a class instance, a holder object.
	Other Type = ""
)

// Expression is one node of an expression tree. Its text is Parts in order, each either literal text
// or a child expression, so a child can be replaced and the whole printed again.
type Expression struct {
	Type  Type
	Parts []Part
}

// Part is literal text or a child expression.
type Part struct {
	Text       string
	Expression *Expression
}

// Statement is one statement: literal text, child expressions and nested blocks, in order.
type Statement struct {
	Parts []StatementPart
}

// StatementPart is literal text, an expression, or a block printed indented between braces.
type StatementPart struct {
	Text       string
	Expression *Expression
	Block      *Block
}

// Block is a list of statements.
type Block struct {
	Statements []*Statement
}

// Program is a whole generated program: the order of its statements is the order it runs in.
// Refusal names an expected failure of shareability, purity or task ownership.
// Part 1 matches a substring; moves match the complete message and RefusalFix.
type Program struct {
	Block   Block
	Refusal string
	// RefusalFix makes a move refusal exact; ordinary parallel refusals remain substrings.
	RefusalFix string
	// Feature records a focused family for reproducing a shrunk finding.
	Feature string
}

// text makes an expression of literal text alone.
func text(t Type, value string) *Expression {
	return &Expression{Type: t, Parts: []Part{{Text: value}}}
}

// compose makes an expression from a format where each @e takes the next child in order.
func compose(t Type, format string, children ...*Expression) *Expression {
	expression := &Expression{Type: t}
	pieces := strings.Split(format, "@e")
	if len(pieces) != len(children)+1 {
		panic("fuzz: format " + format + " doesn't take as many children as it was given")
	}
	for index, piece := range pieces {
		if piece != "" {
			expression.Parts = append(expression.Parts, Part{Text: piece})
		}
		if index < len(children) {
			expression.Parts = append(expression.Parts, Part{Expression: children[index]})
		}
	}
	return expression
}

// statement makes a statement from a format where @e takes the next expression and @b the next
// block, each in order.
func statement(format string, children ...any) *Statement {
	result := &Statement{}
	rest := format
	next := 0
	for {
		index := strings.IndexByte(rest, '@')
		if index < 0 || index+1 >= len(rest) {
			break
		}
		if rest[index+1] != 'e' && rest[index+1] != 'b' {
			panic("fuzz: format " + format + " has an @ that isn't @e or @b")
		}
		if index > 0 {
			result.Parts = append(result.Parts, StatementPart{Text: rest[:index]})
		}
		switch child := children[next].(type) {
		case *Expression:
			result.Parts = append(result.Parts, StatementPart{Expression: child})
		case *Block:
			result.Parts = append(result.Parts, StatementPart{Block: child})
		default:
			panic("fuzz: a statement's child must be an expression or a block")
		}
		next++
		rest = rest[index+2:]
	}
	if next != len(children) {
		panic("fuzz: format " + format + " doesn't take as many children as it was given")
	}
	if rest != "" {
		result.Parts = append(result.Parts, StatementPart{Text: rest})
	}
	return result
}

// Source prints the program as Adamic source.
func (p *Program) Source() string {
	var builder strings.Builder
	for _, statement := range p.Block.Statements {
		statement.write(&builder, 0)
	}
	return builder.String()
}

func (s *Statement) write(builder *strings.Builder, depth int) {
	builder.WriteString(strings.Repeat("\t", depth))
	for _, part := range s.Parts {
		switch {
		case part.Expression != nil:
			part.Expression.write(builder)
		case part.Block != nil:
			builder.WriteString("{\n")
			for _, inner := range part.Block.Statements {
				inner.write(builder, depth+1)
			}
			builder.WriteString(strings.Repeat("\t", depth))
			builder.WriteString("}")
		default:
			builder.WriteString(part.Text)
		}
	}
	builder.WriteString("\n")
}

func (e *Expression) write(builder *strings.Builder) {
	for _, part := range e.Parts {
		if part.Expression != nil {
			part.Expression.write(builder)
		} else {
			builder.WriteString(part.Text)
		}
	}
}

// String prints one expression.
func (e *Expression) String() string {
	var builder strings.Builder
	e.write(&builder)
	return builder.String()
}

// Clone copies a program deeply, so the shrinker can change a candidate without touching the original.
func (p *Program) Clone() *Program {
	return &Program{Block: *p.Block.clone(), Refusal: p.Refusal, RefusalFix: p.RefusalFix, Feature: p.Feature}
}

func (b *Block) clone() *Block {
	result := &Block{Statements: make([]*Statement, len(b.Statements))}
	for index, statement := range b.Statements {
		result.Statements[index] = statement.clone()
	}
	return result
}

func (s *Statement) clone() *Statement {
	result := &Statement{Parts: make([]StatementPart, len(s.Parts))}
	for index, part := range s.Parts {
		result.Parts[index] = StatementPart{Text: part.Text}
		if part.Expression != nil {
			result.Parts[index].Expression = part.Expression.clone()
		}
		if part.Block != nil {
			result.Parts[index].Block = part.Block.clone()
		}
	}
	return result
}

func (e *Expression) clone() *Expression {
	result := &Expression{Type: e.Type, Parts: make([]Part, len(e.Parts))}
	for index, part := range e.Parts {
		result.Parts[index] = Part{Text: part.Text}
		if part.Expression != nil {
			result.Parts[index].Expression = part.Expression.clone()
		}
	}
	return result
}
