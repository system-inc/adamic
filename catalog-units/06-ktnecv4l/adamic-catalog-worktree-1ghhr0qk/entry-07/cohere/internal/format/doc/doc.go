// Package doc is Prettier's document IR and layout algorithm, ported to Go.
//
// Every native printer in cohere turns an AST into a Doc, and this package turns a Doc into text that
// fits a print width. It is the part of Prettier that decides where lines break, and the reason the
// replacement had to be a port rather than typescript-go's formatter: that one has no width at all.
//
// # A port, not a reimplementation
//
// The source of truth is the fork, src/document in system-inc/prettier, and the code here follows it
// statement for statement, including the parts that look accidental. Several are observable and have
// to be copied exactly to stay byte-identical:
//
//   - An IfBreak keyed to a group that has not been printed yet prints nothing at all in the printer,
//     while fits() treats the same case as flat. Two answers to one question, and both are upstream.
//   - The printer advances its column for a string only when more commands remain after it.
//   - fits() counts a pending space only when non-empty text follows, so an empty string between a
//     flat line and a break changes the answer exactly at the width boundary.
//   - Indentation length counts UTF-16 code units, the way JavaScript's String.length does, while text
//     width goes through StringWidth. Go's len() counts bytes, which is neither.
//
// Two upstream details are ported for fidelity but are not observable in the output, which mutation
// showed rather than reasoning: settling the print buffer on every trim (it exists for cursor positions,
// and settled text always ends in a non-space, so no later trim could reach it anyway), and visiting a
// shared group once in PropagateBreaks (the first visit already sets every break the second could).
// Both are kept so the code still reads against its source.
//
// Where the upstream file is named in a comment, read it before changing the Go.
//
// # Identity matters
//
// Prettier's docs are JavaScript objects, and propagateBreaks relies on object identity: a group
// shared by two parents is visited once and mutated in place. So Group is always used through a
// pointer, and a group's identity is its pointer. Copying a Group value would silently fork it.
package doc

// Doc is one node of the document. The concrete types below are the whole language.
type Doc interface {
	isDoc()
}

// Text is a literal string. Empty text is valid and prints nothing, and it does not flush a pending
// space in fits(), which is an upstream detail that matters.
type Text string

// Concat is a sequence, Prettier's array doc.
type Concat []Doc

// Sequence is Prettier's array doc behind a pointer, so it is a Doc without the allocation that putting
// a slice in an interface costs. A printer that cuts its docs from a format's memory builds these, and
// every walk treats one exactly as the Concat of its parts, uncached by identity as a Concat is
// (#fyw36kf).
type Sequence struct{ Parts []Doc }

// Parts is the parts of Prettier's array doc, a Concat or a *Sequence, and whether document is one.
// Nothing else asks a doc whether it is an array: a test fails on a type switch or assertion naming
// either type outside this function, so no walk can know one and miss the other.
func Parts(document Doc) ([]Doc, bool) {
	switch typed := document.(type) {
	case Concat:
		return typed, true
	case *Sequence:
		return typed.Parts, true
	}
	return nil, false
}

// Indent increases indentation by one level for its contents.
type Indent struct{ Contents Doc }

// AlignKind distinguishes Prettier's three non-numeric align arguments from a width.
type AlignKind int

const (
	// AlignWidth adds Width columns of alignment, or dedents one level when Width is negative.
	AlignWidth AlignKind = iota
	// AlignString adds a literal string, used by markdown list markers and the like.
	AlignString
	// AlignRoot marks the current indentation as the root that literal lines return to.
	AlignRoot
	// AlignDedentToRoot returns to the marked root, Prettier's Number.NEGATIVE_INFINITY.
	AlignDedentToRoot
)

// Align is Prettier's align(n, contents) with n spelled out as a kind.
type Align struct {
	Kind     AlignKind
	Width    int
	String   string
	Contents Doc
}

// trimDoc is the trim builder's single instance.
type trimDoc struct{}

// Group lays out flat if it fits, broken otherwise.
//
// Always used as *Group: identity is the pointer, and propagateBreaks writes Break in place.
type Group struct {
	ID             *GroupID
	Contents       Doc
	Break          bool
	ExpandedStates []Doc
}

// GroupID names a group so IfBreak and IndentIfBreak can ask how it was printed.
//
// A pointer to an empty struct, so two ids are equal only if they are the same id, the way two
// JavaScript Symbols are.
type GroupID struct{ name string }

// NewGroupID mints an id. The name is for debugging only and never compared.
func NewGroupID(name string) *GroupID { return &GroupID{name: name} }

// Fill packs content onto lines as tightly as it fits. Parts alternate content and separator.
type Fill struct{ Parts []Doc }

// IfBreak chooses between two docs by whether a group broke.
type IfBreak struct {
	BreakContents Doc
	FlatContents  Doc
	GroupID       *GroupID
}

// IndentIfBreak indents its contents only when the named group broke, or only when it did not.
type IndentIfBreak struct {
	Contents Doc
	GroupID  *GroupID
	Negate   bool
}

// LineSuffix defers its contents to the end of the line, used for trailing comments.
type LineSuffix struct{ Contents Doc }

// lineSuffixBoundaryDoc forces deferred line suffixes out before it.
type lineSuffixBoundaryDoc struct{}

// Line is a possible line break.
type Line struct {
	Soft    bool
	Hard    bool
	Literal bool
}

// Label tags a doc for printer heuristics without changing how it prints.
type Label struct {
	Label    string
	Contents Doc
}

// breakParentDoc forces every enclosing group to break.
type breakParentDoc struct{}

func (Text) isDoc()                  {}
func (Concat) isDoc()                {}
func (*Sequence) isDoc()             {}
func (*Indent) isDoc()               {}
func (*Align) isDoc()                {}
func (trimDoc) isDoc()               {}
func (*Group) isDoc()                {}
func (*Fill) isDoc()                 {}
func (*IfBreak) isDoc()              {}
func (*IndentIfBreak) isDoc()        {}
func (*LineSuffix) isDoc()           {}
func (lineSuffixBoundaryDoc) isDoc() {}
func (*Line) isDoc()                 {}
func (*Label) isDoc()                {}
func (breakParentDoc) isDoc()        {}

// The builders, named as upstream names them so a printer port reads like its source.
var (
	// Trim removes trailing whitespace on the current line.
	Trim Doc = trimDoc{}
	// LineSuffixBoundary flushes deferred line suffixes.
	LineSuffixBoundary Doc = lineSuffixBoundaryDoc{}
	// BreakParent forces enclosing groups to break.
	BreakParent Doc = breakParentDoc{}

	// LineDoc is a space when flat and a newline when broken.
	LineDoc Doc = &Line{}
	// Softline is nothing when flat and a newline when broken.
	Softline Doc = &Line{Soft: true}
	// HardlineWithoutBreakParent always breaks but does not force its parent to.
	HardlineWithoutBreakParent Doc = &Line{Hard: true}
	// LiterallineWithoutBreakParent breaks to the root indentation without breaking its parent.
	LiterallineWithoutBreakParent Doc = &Line{Hard: true, Literal: true}
	// Hardline always breaks, and breaks every enclosing group.
	Hardline Doc = Concat{HardlineWithoutBreakParent, BreakParent}
	// Literalline breaks to the root indentation and breaks every enclosing group.
	Literalline Doc = Concat{LiterallineWithoutBreakParent, BreakParent}
)

// GroupOptions are upstream's group options.
type GroupOptions struct {
	ID             *GroupID
	ShouldBreak    bool
	ExpandedStates []Doc
}

// NewGroup is upstream's group().
func NewGroup(contents Doc, options GroupOptions) *Group {
	return &Group{ID: options.ID, Contents: contents, Break: options.ShouldBreak, ExpandedStates: options.ExpandedStates}
}

// ConditionalGroup is upstream's conditionalGroup(): the first state is the contents.
func ConditionalGroup(states []Doc, options GroupOptions) *Group {
	options.ExpandedStates = states
	return NewGroup(states[0], options)
}

// NewIndent is upstream's indent().
func NewIndent(contents Doc) Doc { return &Indent{Contents: contents} }

// NewAlign is upstream's align(n) for a numeric n.
func NewAlign(width int, contents Doc) Doc {
	return &Align{Kind: AlignWidth, Width: width, Contents: contents}
}

// AlignWithString is upstream's align(n) for a string n.
func AlignWithString(prefix string, contents Doc) Doc {
	return &Align{Kind: AlignString, String: prefix, Contents: contents}
}

// Dedent is upstream's dedent(): align(-1).
func Dedent(contents Doc) Doc { return NewAlign(-1, contents) }

// DedentToRoot is upstream's dedentToRoot(): align(-Infinity).
func DedentToRoot(contents Doc) Doc { return &Align{Kind: AlignDedentToRoot, Contents: contents} }

// MarkAsRoot is upstream's markAsRoot(): align({type: "root"}).
func MarkAsRoot(contents Doc) Doc { return &Align{Kind: AlignRoot, Contents: contents} }

// NewFill is upstream's fill().
func NewFill(parts []Doc) Doc { return &Fill{Parts: parts} }

// NewIfBreak is upstream's ifBreak(). Upstream defaults flatContents to "", and so does a nil here.
func NewIfBreak(breakContents Doc, flatContents Doc, groupID *GroupID) Doc {
	if flatContents == nil {
		flatContents = Text("")
	}
	return &IfBreak{BreakContents: breakContents, FlatContents: flatContents, GroupID: groupID}
}

// NewIndentIfBreak is upstream's indentIfBreak().
func NewIndentIfBreak(contents Doc, groupID *GroupID, negate bool) Doc {
	return &IndentIfBreak{Contents: contents, GroupID: groupID, Negate: negate}
}

// NewLineSuffix is upstream's lineSuffix().
func NewLineSuffix(contents Doc) Doc { return &LineSuffix{Contents: contents} }

// NewLabel is upstream's label(): an empty label returns the contents unwrapped, as upstream does
// for a falsy label.
func NewLabel(label string, contents Doc) Doc {
	if label == "" {
		return contents
	}
	return &Label{Label: label, Contents: contents}
}

// Join is upstream's join(): separator between each part. Upstream's is an array a caller may spread,
// so this is a Concat, typed as one.
func Join(separator Doc, parts []Doc) Concat {
	result := make(Concat, 0, 2*len(parts))
	for index, part := range parts {
		if index > 0 {
			result = append(result, separator)
		}
		result = append(result, part)
	}
	return result
}
