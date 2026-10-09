// The instruction set.
//
// 43 variants, matching upstream's count and names. See high_level_intermediate_representation.go's package comment for why fidelity
// to upstream is the priority here and where it is deliberately broken.
//
// # Why an interface rather than a tagged struct
//
// Go has no sum type. The two candidates were a struct with a `Kind` field and a union of optional
// payload fields, or an interface with one implementing struct per variant.
//
// The interface wins on the property that matters most for a set this size: a variant's fields are
// only reachable after the type switch has established which variant it is. The tagged-struct
// alternative makes every field reachable on every variant, so `value.Callee` compiles on a
// `Primitive` and yields a zero value at runtime. With 43 variants and a dozen passes to be written
// against them, that is a class of bug the compiler should be catching.
//
// The cost is that Go cannot check a type switch for exhaustiveness. That is mitigated by
// `instructionValue()` being unexported, which closes the set to this package, and by `Visit` in
// visitor.go being the single place a walk over all variants is written: a new variant added
// without a case there fails the round-trip tests rather than silently doing nothing.
package high_level_intermediate_representation

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// InstructionValue is what an instruction computes.
//
// The set is closed: the unexported marker method means only this package can add a variant.
type InstructionValue interface {
	instructionValue()
}

// ---------------------------------------------------------------------------
// Loads and stores
// ---------------------------------------------------------------------------

// LoadLocal reads a local binding.
type LoadLocal struct{ Place Place }

// LoadContext reads a binding captured from an enclosing function.
//
// Distinct from LoadLocal because the value can change between reads without anything in this
// function storing to it, which every dataflow analysis over this IR must respect.
type LoadContext struct{ Place Place }

// DeclareLocal introduces a binding with no initializer.
type DeclareLocal struct {
	LValue Place
	Kind   InstructionKind
}

// DeclareContext introduces a binding that a nested function will capture.
type DeclareContext struct {
	LValue Place
	Kind   InstructionKind
}

// StoreLocal writes to a local binding.
type StoreLocal struct {
	LValue Place
	Value  Place
	Kind   InstructionKind
}

// StoreContext writes to a binding captured from an enclosing function.
type StoreContext struct {
	LValue Place
	Value  Place
	Kind   InstructionKind
}

// Destructure binds the parts of a value to a pattern.
//
// The pattern is kept structured rather than expanded into property loads because the distinction
// between `const {a} = o` and `const a = o.a` is observable: the first throws on a nullish `o` at
// the destructure, and an effect pass wants one instruction to attribute that to.
type Destructure struct {
	LValue  Pattern
	Value   Place
	Kind    InstructionKind
	Pattern Pattern
}

// LoadGlobal reads a global binding.
type LoadGlobal struct {
	Name        string
	BindingKind GlobalBindingKind
	// Source is the module a value was imported from, empty for a true global.
	Source string
	// Imported is the name as exported by Source, empty for a default or namespace import.
	Imported string
}

// StoreGlobal writes to a global binding.
//
// This instruction is the entire detection for the `globals` rule: a write to a global is exactly
// this instruction existing. That the rule sits in a later tier upstream is an artifact of where
// the diagnostic is emitted, not of what it needs.
type StoreGlobal struct {
	Name  string
	Value Place
}

// GlobalBindingKind is how a global name came to be in scope.
type GlobalBindingKind uint8

const (
	// GlobalBindingKindGlobal is a true global with no declaration in the program.
	GlobalBindingKindGlobal GlobalBindingKind = iota
	// GlobalBindingKindModuleLocal is a module-scope binding declared in this file.
	GlobalBindingKindModuleLocal
	// GlobalBindingKindImportDefault is a default import.
	GlobalBindingKindImportDefault
	// GlobalBindingKindImportNamespace is a namespace import.
	GlobalBindingKindImportNamespace
	// GlobalBindingKindImportSpecifier is a named import.
	GlobalBindingKindImportSpecifier
)

// ---------------------------------------------------------------------------
// Property access
// ---------------------------------------------------------------------------

// PropertyLoad reads a statically named property.
type PropertyLoad struct {
	Object   Place
	Property string
	// Optional records that this came from `a?.b`, so a nullish object yields undefined rather
	// than throwing. The branch itself is a terminal; this flag is what tells a pass reading one
	// instruction that the load cannot throw.
	Optional bool
}

// PropertyStore writes a statically named property.
type PropertyStore struct {
	Object   Place
	Property string
	Value    Place
}

// PropertyDelete deletes a statically named property.
type PropertyDelete struct {
	Object   Place
	Property string
}

// ComputedLoad reads a dynamically named property.
type ComputedLoad struct {
	Object   Place
	Property Place
	Optional bool
}

// ComputedStore writes a dynamically named property.
type ComputedStore struct {
	Object   Place
	Property Place
	Value    Place
}

// ComputedDelete deletes a dynamically named property.
type ComputedDelete struct {
	Object   Place
	Property Place
}

// ---------------------------------------------------------------------------
// Calls
// ---------------------------------------------------------------------------

// CallExpression calls a value.
type CallExpression struct {
	Callee       Place
	Args         []Argument
	Optional     bool
	CalleeOrigin ModuleExportOrigin
}

// MethodCall calls a property of an object, keeping the receiver.
//
// Separate from CallExpression because the receiver is the `this` of the call, and an effect pass
// must attribute mutation to it. Lowering a method call as a property load plus a call would lose
// that the loaded function and the object are related.
type MethodCall struct {
	Receiver     Place
	Property     Place
	Args         []Argument
	Optional     bool
	CalleeOrigin ModuleExportOrigin
}

type ModuleExportOrigin struct {
	Module string
	Export string
}

// NewExpression constructs a value.
type NewExpression struct {
	Callee Place
	Args   []Argument
}

// Argument is one argument to a call: a value, or a spread of one.
type Argument struct {
	Place  Place
	Spread bool
}

// ---------------------------------------------------------------------------
// Operators and literals
// ---------------------------------------------------------------------------

// BinaryExpression applies a binary operator.
//
// `&&`, `||`, and `??` are NOT here. They short-circuit, so they are control flow and lower to a
// Logical terminal.
type BinaryExpression struct {
	Left     Place
	Operator string
	Right    Place
}

// UnaryExpression applies a unary operator.
type UnaryExpression struct {
	Operator string
	Value    Place
}

// PrefixUpdate is `++x` or `--x`: the value after the update.
type PrefixUpdate struct {
	LValue    Place
	Operation string
	Value     Place
}

// PostfixUpdate is `x++` or `x--`: the value before the update.
type PostfixUpdate struct {
	LValue    Place
	Operation string
	Value     Place
}

// Primitive is a literal number, string, boolean, null, or undefined.
type Primitive struct{ Value any }

// RegExpLiteral is a regular expression literal.
type RegExpLiteral struct {
	Pattern string
	Flags   string
}

// TemplateLiteral is a template string with its interpolations.
type TemplateLiteral struct {
	Quasis   []string
	Subexprs []Place
}

// TaggedTemplateExpression applies a tag function to a template literal.
type TaggedTemplateExpression struct {
	Tag      Place
	Quasis   []string
	Subexprs []Place
}

// TypeCastExpression is `x as T`, `<T>x`, or a Flow cast: a no-op at runtime carrying a type.
type TypeCastExpression struct {
	Value Place
	// Node is the type annotation, handed to the checker rather than interpreted here.
	Node *ast.Node
}

// MetaProperty is `import.meta` or `new.target`.
type MetaProperty struct {
	Meta     string
	Property string
}

// ---------------------------------------------------------------------------
// Aggregates
// ---------------------------------------------------------------------------

// ObjectExpression constructs an object.
type ObjectExpression struct{ Properties []ObjectProperty }

// ObjectProperty is one entry in an object literal: a keyed value, a spread, or a method.
type ObjectProperty struct {
	// Key is the property name. Empty when Spread is set.
	Key string
	// ComputedKey is the key expression for `{[k]: v}`. Set instead of Key.
	ComputedKey *Place
	Value       Place
	Spread      bool
}

// ObjectMethod is a method defined in an object literal.
type ObjectMethod struct {
	Key      string
	Function FunctionId
}

// ArrayExpression constructs an array.
type ArrayExpression struct{ Elements []ArrayElement }

// ArrayElement is one element: a value, a spread, or a hole from `[a, , b]`.
type ArrayElement struct {
	Place  Place
	Spread bool
	Hole   bool
}

// FunctionExpression is a function or arrow expression.
//
// The body is in Function.Functions, addressed by id, so a pass that does not descend into nested
// functions never walks it.
type FunctionExpression struct {
	Function FunctionId
	// Captures are the values from this function that the nested one closes over.
	Captures []Place
}

// ---------------------------------------------------------------------------
// Iteration and async
// ---------------------------------------------------------------------------

// Await suspends on a value.
type Await struct{ Value Place }

// GetIterator obtains an iterator from an iterable, the first step of `for...of`.
type GetIterator struct{ Value Place }

// IteratorNext advances an iterator.
type IteratorNext struct {
	Iterator   Place
	Collection Place
}

// NextPropertyOf yields the next key in a `for...in`.
type NextPropertyOf struct{ Value Place }

// ---------------------------------------------------------------------------
// JSX
// ---------------------------------------------------------------------------
//
// Three variants for syntax the grammar has. See the package comment for why these are in the core
// set while the reactive-scope terminals are not.

// JsxExpression is a JSX element.
type JsxExpression struct {
	// Tag is the element name. For a lowercase host element this is the literal string; for a
	// component it is a Place naming the component value.
	Tag      JsxTag
	Props    []JsxAttribute
	Children []Place
}

// JsxTag is either a host element name or a component value.
type JsxTag struct {
	// Name is set for a host element such as `div`.
	Name string
	// Place is set for a component reference such as `<Foo />`.
	Place *Place
}

// JsxAttribute is one prop: a named value or a spread.
type JsxAttribute struct {
	Name   string
	Value  Place
	Spread bool
}

// JsxFragment is `<>...</>`.
type JsxFragment struct{ Children []Place }

// JsxText is literal text inside JSX.
type JsxText struct{ Value string }

// ---------------------------------------------------------------------------
// Memoization markers
// ---------------------------------------------------------------------------
//
// Inserted by a later pass and read by a later pass. Nothing in lowering produces them and no
// generic pass must interpret them; they cost one variant each.

// ManualMemoDependency is one entry a developer WROTE in a `useMemo`/`useCallback` dependency
// array: an access path, not a value.
//
// This is upstream's `ManualMemoDependency` and it is deliberately the same shape as
// `ReactiveScopeDependency` in `dependencies.go`, because the entire purpose of these markers is
// for the two to be COMPARED. `preserve-manual-memoization` reports when what the developer wrote
// and what the compiler inferred disagree, so a difference in spelling between the two sides would
// become a false report at every call site.
//
// `Root` is the base of the path. A dependency array entry can be rooted at a local binding
// (`props` in `[props.items]`) or at a global, and the two are not interchangeable: a global has no
// identifier in this function to compare against, so it is carried by name.
//
// The path reuses `DependencyPathEntry` rather than declaring a second spelling of the same idea.
// That reuse is load-bearing rather than tidy: `DependencyPathEntry.Optional` is compared by
// `equalPaths`, so `props.a` and `props?.a` are different dependencies on BOTH sides of the
// comparison. Collapsing them on this side only would make a one-character source difference
// invisible to the rule whose whole job is to see it.
type ManualMemoDependency struct {
	Root ManualMemoRoot
	Path []DependencyPathEntry
}

// ManualMemoRoot is the base of a manual dependency path: a local binding, or a global by name.
//
// Exactly one field is meaningful, selected by `IsGlobal`. A struct rather than an interface
// because this is carried inside an instruction value and every pass that walks operands must be
// able to see the `Place` without a type switch.
type ManualMemoRoot struct {
	// IsGlobal selects which of the two fields below is meaningful.
	IsGlobal bool
	// Place is the local binding the path is rooted at, meaningful when IsGlobal is false.
	Place Place
	// Name is the global's name, meaningful when IsGlobal is true.
	Name string
}

// StartMemoize marks the beginning of a manually memoized region.
//
// `Deps` is nil when the call had NO dependency array at all, which is a different fact from an
// empty array and is why the field is a nil-able slice rather than a length. `useMemo(fn)` with no
// second argument recomputes every render and is not a memoization claim; `useMemo(fn, [])` claims
// the value never changes. A rule comparing declared against inferred dependencies must not read
// the first as the second.
type StartMemoize struct {
	// ManualMemoId pairs this marker with its `FinishMemoize`.
	//
	// Necessary rather than decorative: memo calls nest (a `useMemo` whose callback body contains
	// another), so the markers do not form a simple stack in instruction order once the callbacks
	// are inlined, and a consumer matching by proximity would pair the wrong two.
	ManualMemoId int
	// Deps are the dependencies the developer WROTE, nil when the dependency array was absent.
	Deps []ManualMemoDependency
}

// FinishMemoize marks the end of a manually memoized region.
type FinishMemoize struct {
	// ManualMemoId pairs this marker with its `StartMemoize`.
	ManualMemoId int
	// Value is the memoized declaration: the call's result for `useMemo`, the callback itself for
	// `useCallback`. Upstream names this field `decl`.
	Value Place
	// Pruned records that the memoization was discarded.
	Pruned bool
}

// ---------------------------------------------------------------------------
// Escapes
// ---------------------------------------------------------------------------

// Debugger is a `debugger` statement.
type Debugger struct{}

// UnsupportedNode is syntax lowering does not model.
//
// This variant is why lowering never fails on a construct it does not know. A rule that must refuse
// such a function checks for it; every other pass treats it as an opaque value with unknown
// effects, which is the conservative answer.
type UnsupportedNode struct {
	Node *ast.Node
	// Reason is a short phrase naming the construct, for diagnostics.
	Reason string
}

// ---------------------------------------------------------------------------

func (*LoadLocal) instructionValue()                {}
func (*LoadContext) instructionValue()              {}
func (*DeclareLocal) instructionValue()             {}
func (*DeclareContext) instructionValue()           {}
func (*StoreLocal) instructionValue()               {}
func (*StoreContext) instructionValue()             {}
func (*Destructure) instructionValue()              {}
func (*LoadGlobal) instructionValue()               {}
func (*StoreGlobal) instructionValue()              {}
func (*PropertyLoad) instructionValue()             {}
func (*PropertyStore) instructionValue()            {}
func (*PropertyDelete) instructionValue()           {}
func (*ComputedLoad) instructionValue()             {}
func (*ComputedStore) instructionValue()            {}
func (*ComputedDelete) instructionValue()           {}
func (*CallExpression) instructionValue()           {}
func (*MethodCall) instructionValue()               {}
func (*NewExpression) instructionValue()            {}
func (*BinaryExpression) instructionValue()         {}
func (*UnaryExpression) instructionValue()          {}
func (*PrefixUpdate) instructionValue()             {}
func (*PostfixUpdate) instructionValue()            {}
func (*Primitive) instructionValue()                {}
func (*RegExpLiteral) instructionValue()            {}
func (*TemplateLiteral) instructionValue()          {}
func (*TaggedTemplateExpression) instructionValue() {}
func (*TypeCastExpression) instructionValue()       {}
func (*MetaProperty) instructionValue()             {}
func (*ObjectExpression) instructionValue()         {}
func (*ObjectMethod) instructionValue()             {}
func (*ArrayExpression) instructionValue()          {}
func (*FunctionExpression) instructionValue()       {}
func (*Await) instructionValue()                    {}
func (*GetIterator) instructionValue()              {}
func (*IteratorNext) instructionValue()             {}
func (*NextPropertyOf) instructionValue()           {}
func (*JsxExpression) instructionValue()            {}
func (*JsxFragment) instructionValue()              {}
func (*JsxText) instructionValue()                  {}
func (*StartMemoize) instructionValue()             {}
func (*FinishMemoize) instructionValue()            {}
func (*Debugger) instructionValue()                 {}
func (*UnsupportedNode) instructionValue()          {}
