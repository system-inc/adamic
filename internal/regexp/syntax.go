// Package regexp parses and executes ECMAScript regular expression patterns.
// The compiler preserves JavaScript backtracking semantics in portable bytecode.
package regexp

import "math/big"

// Flags are the flags accepted by ECMAScript 2025.
type Flags struct {
	HasIndices, Global, IgnoreCase, Multiline bool
	DotAll, Unicode, UnicodeSets, Sticky      bool
}

// Node is implemented by every syntax tree node.
type Node interface{ regexpNode() }

// Pattern is a parsed regular expression.
type Pattern struct {
	Flags Flags
	Body  *Disjunction
}

// Disjunction contains alternatives separated by |.
type Disjunction struct{ Alternatives []*Alternative }

// Alternative is a sequence of terms.
type Alternative struct{ Terms []Node }

// Character is a literal UTF-16 code point or an escape. Raw is its source spelling.
type Character struct {
	Value rune
	Raw   string
	Kind  CharacterKind
}

type CharacterKind uint8

const (
	Literal CharacterKind = iota
	Escape
	ClassEscape
	PropertyEscape
)

// Assertion is a zero-width assertion.
type Assertion struct{ Kind AssertionKind }

type AssertionKind uint8

const (
	Start AssertionKind = iota
	End
	WordBoundary
	NotWordBoundary
)

// Dot matches any character subject to the flags.
type Dot struct{}

// Quantifier applies arbitrary-precision ECMAScript integer bounds to Atom.
// Max is nil for an unbounded quantifier.
type Quantifier struct {
	Atom     Node
	Min, Max *big.Int
	Greedy   bool
}

// Group includes capturing, lookaround and ordinary non-capturing groups.
type Group struct {
	Kind    GroupKind
	Name    string
	Body    *Disjunction
	Enable  Flags
	Disable Flags
}

type GroupKind uint8

const (
	Capturing GroupKind = iota
	NonCapturing
	PositiveLookahead
	NegativeLookahead
	PositiveLookbehind
	NegativeLookbehind
)

// Backreference refers to a capture by number or name.
type Backreference struct {
	Index int
	Name  string
}

// CharacterClass is a traditional class or a Unicode-sets class expression.
type CharacterClass struct {
	Negated bool
	Expr    ClassExpression
}

type ClassExpression interface{ classExpression() }

type ClassNegation struct{ Operand ClassExpression }
type ClassUnion struct{ Operands []ClassExpression }
type ClassIntersection struct{ Left, Right ClassExpression }
type ClassSubtraction struct{ Left, Right ClassExpression }
type ClassRange struct{ From, To *Character }
type ClassCharacter struct{ Character *Character }

// ClassString is a \q{...} string-disjunction. Alternatives may contain an
// empty string, as ECMAScript permits.
type ClassString struct{ Alternatives [][]*Character }

func (*Disjunction) regexpNode()            {}
func (*Character) regexpNode()              {}
func (*Assertion) regexpNode()              {}
func (*Dot) regexpNode()                    {}
func (*Quantifier) regexpNode()             {}
func (*Group) regexpNode()                  {}
func (*Backreference) regexpNode()          {}
func (*CharacterClass) regexpNode()         {}
func (*ClassNegation) classExpression()     {}
func (*ClassUnion) classExpression()        {}
func (*ClassIntersection) classExpression() {}
func (*ClassSubtraction) classExpression()  {}
func (*ClassRange) classExpression()        {}
func (*ClassCharacter) classExpression()    {}
func (*ClassString) classExpression()       {}
