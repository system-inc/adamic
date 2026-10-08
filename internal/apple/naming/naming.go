// Package naming assigns Apple declarations stable Adamic names without an SDK.
// Name handles one declaration; Build validates the complete export surface before
// a generator emits anything. Never emit a surface that Build rejects.
package naming

import (
	"fmt"
	"sort"
	"strings"
)

type Kind string

const (
	Class          Kind = "class"
	Protocol       Kind = "protocol"
	Category       Kind = "category"
	InstanceMethod Kind = "instance-method"
	ClassMethod    Kind = "class-method"
	Property       Kind = "property"
	Enum           Kind = "enum"
	EnumConstant   Kind = "enum-constant"
	OptionSet      Kind = "option-set"
	OptionConstant Kind = "option-constant"
	Struct         Kind = "struct"
	StructField    Kind = "struct-field"
	CFunction      Kind = "c-function"
	Constant       Kind = "constant"
	Typedef        Kind = "typedef"
)

type Nullability string

const (
	Unspecified Nullability = ""
	Nonnull     Nullability = "nonnull"
	Nullable    Nullability = "nullable"
)

// Type records the written spelling, not an Adamic or Swift spelling. Element
// supplies a lightweight generic collection's element type for word omission.
// OptionCases supplies the option set's complete literal domain.
type Type struct {
	Framework   string // declaring framework for reserved global type names
	Spelling    string
	Nullability Nullability
	Element     string
	OptionCases []string
	Object      bool          // clang says this written pointer points to an Objective-C object
	Reference   bool          // pointer-like typedef, known from clang, even without a written star
	Function    *FunctionType // block/function type facts, without parsing C declarators
}
type FunctionType struct {
	Parameters []Type
	Result     Type
}
type Parameter struct {
	Name string
	Type Type
	// DefaultArgument is Swift's inferred default-argument fact. The generator
	// must supply it when applicable; nullability alone does not imply a default.
	DefaultArgument bool
}
type Enumerator struct {
	Name        string
	SwiftName   string
	Unavailable bool
	Deprecated  bool
}

// Parent is the original lexical parent, including an enum's owning class when
// nested. ParentSwiftName is its imported name when API notes rename it.
type Declaration struct {
	Kind            Kind
	Name            string
	Parent          string
	ParentSwiftName string
	Framework       string
	Parameters      []Parameter
	Result          Type
	SwiftName       string       // NS_SWIFT_NAME, exactly as written.
	APIName         string       // effective SDK API-note/CF-bridge name, if there is no SwiftName.
	RefinedForSwift bool         // NS_REFINED_FOR_SWIFT; retained, never silently hidden.
	Enumerators     []Enumerator // all siblings, required for inferred case names
	PropertyNames   []string     // includes inherited properties, for omission protection
	// TakenNames are names this member may not take, of the same staticness: the owner's and its
	// ancestors' properties, and inherited methods it doesn't override. A method whose name would be
	// one of them gives way (see foldFirstLabel).
	TakenNames    []string
	ClassProperty bool // a property of the class, not its instances (@property (class))
	// KeepNeedlessWords is Swift's conflict rule: a method whose omitted name would collide with
	// a sibling's keeps its words (NSArrayController's addObject: beside add:).
	KeepNeedlessWords bool
	Getter            string // a property's getter selector, including a BOOL is-prefixed getter
	Setter            string
	// AccessorProperty joins getter/setter declarations to a property. Such
	// accessors are emitted through the property, not independently by Build.
	AccessorProperty string
}

type Argument struct {
	Index      int // original Objective-C parameter index, for marshalling
	SwiftLabel string
	Name       string
	Type       string
}
type Layout struct {
	Constructor bool
	Positional  []Argument
	Options     []Argument // one trailing object, in original parameter order
}
type Output struct {
	Name            string
	Module          string
	Original        string
	SwiftName       string
	RefinedForSwift bool
	Type            string
	Arguments       *Layout
	Getter          string
	Setter          string
}

func isType(k Kind) bool {
	switch k {
	case Class, Protocol, Category, Enum, OptionSet, Struct, Typedef:
		return true
	}
	return false
}
func isCallable(k Kind) bool { return k == InstanceMethod || k == ClassMethod || k == CFunction }
func isCase(k Kind) bool     { return k == EnumConstant || k == OptionConstant }

// Name applies the ordered rules. Errors mean that the generator needs more
// header facts or that this declaration cannot be represented safely yet.
func Name(d Declaration) (Output, error) {
	switch d.Kind {
	case Class, Protocol, Category, InstanceMethod, ClassMethod, Property, Enum, EnumConstant, OptionSet, OptionConstant, Struct, StructField, CFunction, Constant, Typedef:
	default:
		return Output{}, fmt.Errorf("unknown declaration kind %q", d.Kind)
	}
	if !identifier(d.Framework) || d.Name == "" {
		return Output{}, fmt.Errorf("missing or invalid framework/name")
	}
	if (d.Kind == InstanceMethod || d.Kind == ClassMethod || d.Kind == Property || d.Kind == StructField || d.Kind == Category || isCase(d.Kind)) && d.Parent == "" {
		return Output{}, fmt.Errorf("%s %q needs a parent", d.Kind, d.Name)
	}
	for _, context := range []string{d.Parent, exportContext(d)} {
		if context != "" {
			for _, component := range strings.Split(context, ".") {
				if !identifier(component) {
					return Output{}, fmt.Errorf("invalid parent/context %q", context)
				}
			}
		}
	}
	if !isCallable(d.Kind) && !identifier(d.Name) {
		return Output{}, fmt.Errorf("invalid declaration name %q", d.Name)
	}
	swift, base, labels, err := swiftName(d)
	if err != nil {
		return Output{}, err
	}
	result := Output{Original: d.Name, SwiftName: swift, RefinedForSwift: d.RefinedForSwift, Getter: d.Getter, Setter: d.Setter}
	result.Name = normalize(base, isType(d.Kind) || isCase(d.Kind))
	if isType(d.Kind) {
		result.Name = typeName(base, d.Framework)
	}
	if result.Name == "" {
		return Output{}, fmt.Errorf("empty imported name for %q", d.Name)
	}
	if isCase(d.Kind) {
		result.Name = "'" + result.Name + "'"
	}
	owner := exportContext(d)
	if owner == "" {
		owner = base
	}
	// Nested types live in their lexical owner's module; their exported name is
	// still the terminal component. The enclosing type owns its members too.
	owner = strings.Split(owner, ".")[0]
	result.Module = "apple/" + strings.ToLower(d.Framework) + "/" + kebab(typeName(owner, d.Framework))
	if isCallable(d.Kind) && d.AccessorProperty == "" {
		layout := &Layout{Constructor: strings.TrimPrefix(base, "__") == "init"}
		if layout.Constructor {
			result.Name = "constructor"
		}
		keys := map[string]bool{}
		for i, p := range d.Parameters {
			typ, e := MapType(p.Type)
			if e != nil {
				return Output{}, fmt.Errorf("%s parameter %d: %w", d.Name, i, e)
			}
			label := labels[i]
			key := normalize(label, false)
			if label == "_" {
				key = normalize(p.Name, false)
				if key == "" {
					key = fmt.Sprintf("argument%d", i+1)
				}
			}
			arg := Argument{Index: i, SwiftLabel: label, Name: key, Type: typ}
			if (i == 0 && label == "_") || d.Kind == CFunction {
				layout.Positional = append(layout.Positional, arg)
			} else {
				if label == "_" {
					return Output{}, fmt.Errorf("%s: later unlabeled argument %d cannot key an options object", d.Name, i)
				}
				if !identifier(key) || keys[key] {
					return Output{}, fmt.Errorf("%s: duplicate or invalid options key %q", d.Name, key)
				}
				keys[key] = true
				layout.Options = append(layout.Options, arg)
			}
		}
		result.Arguments = layout
		if (d.Kind == InstanceMethod || d.Kind == ClassMethod) && !layout.Constructor && contains(d.TakenNames, result.Name) {
			if len(layout.Options) > 0 {
				foldFirstLabel(&result, d)
			} else if !d.KeepNeedlessWords {
				// With no label to fold, Swift's conflict rule: keep the words omission dropped,
				// when that frees the name (NSControl's drawCell: beside NSView's drawRect:, both
				// draw(_:) once their types' words are dropped).
				kept := d
				kept.KeepNeedlessWords = true
				if named, err := Name(kept); err == nil && !contains(d.TakenNames, named.Name) {
					return named, nil
				}
			}
		}
	}
	if d.Result.Spelling != "" {
		resultType := d.Result
		if resultType.Spelling == "instancetype" {
			resultType.Spelling = d.Parent + " *"
			resultType.Object = true
			resultType.Framework = d.Framework
		}
		result.Type, err = MapType(resultType)
		if err != nil {
			return Output{}, err
		}
	}
	return result, nil
}

// foldFirstLabel renames a method that would take a name already taken: Swift tells
// abbreviation(for:) from the property abbreviation, and NSMatrix's cell(atRow:column:) from
// NSControl's cell, by their labels, and a class member can't, so the
// first label joins the name and its argument becomes positional, Objective-C's own reading
// (abbreviationForDate: is abbreviationFor(date), isValidDateInCalendar: is isValidDateIn(calendar)).
func foldFirstLabel(result *Output, d Declaration) {
	layout := result.Arguments
	first := layout.Options[0]
	result.Name += normalize(first.SwiftLabel, true)
	first.Name = normalize(d.Parameters[first.Index].Name, false)
	if first.Name == "" {
		first.Name = fmt.Sprintf("argument%d", first.Index+1)
	}
	layout.Positional = append(layout.Positional, first)
	layout.Options = layout.Options[1:]
}
func contains(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
}

// MapType refuses unknown pointer shapes rather than claiming they are objects.
// Unspecified reference nullability is deliberately nullable. Scalar values
// cannot be null unless explicitly marked nullable.
func MapType(t Type) (string, error) {
	if t.Nullability != Unspecified && t.Nullability != Nonnull && t.Nullability != Nullable {
		return "", fmt.Errorf("invalid nullability %q", t.Nullability)
	}
	s := strings.TrimSpace(t.Spelling)
	for _, q := range []string{"const ", "volatile ", "_Nonnull", "_Nullable", "__nonnull", "__nullable"} {
		s = strings.TrimSpace(strings.ReplaceAll(s, q, ""))
	}
	if t.Function != nil {
		args := make([]string, len(t.Function.Parameters))
		for i, p := range t.Function.Parameters {
			mapped, e := MapType(p)
			if e != nil {
				return "", e
			}
			args[i] = fmt.Sprintf("argument%d: %s", i+1, mapped)
		}
		result, e := MapType(t.Function.Result)
		if e != nil {
			return "", e
		}
		name := "(" + strings.Join(args, ", ") + ") => " + result
		if t.Nullability != Nonnull {
			name = "(" + name + ") | null"
		}
		return name, nil
	}
	pointer := strings.Contains(s, "*")
	base := strings.TrimSpace(strings.TrimSuffix(s, "*"))
	var name string
	switch base {
	case "CGFloat", "NSInteger", "NSUInteger", "double", "float", "int", "unsigned int", "long", "unsigned long", "long long", "unsigned long long", "int64_t", "uint64_t", "short", "unsigned short", "size_t":
		name = "number"
	case "BOOL", "bool", "_Bool":
		name = "boolean"
	case "void":
		name = "void"
	case "id":
		name = "unknown"
		t.Reference = true
	case "instancetype":
		return "", fmt.Errorf("instancetype needs a resolved concrete object type")
	case "NSString":
		if !pointer {
			return "", fmt.Errorf("NSString needs an object pointer")
		}
		name = "string"
	default:
		if !identifier(base) {
			return "", fmt.Errorf("unsupported type spelling %q", t.Spelling)
		}
		if pointer && !t.Object {
			return "", fmt.Errorf("unknown object pointer %q", t.Spelling)
		}
		name = typeName(base, t.Framework)
	}
	if pointer && (name == "number" || name == "boolean" || name == "void") {
		return "", fmt.Errorf("unsupported scalar pointer %q", t.Spelling)
	}
	if len(t.OptionCases) > 0 {
		if pointer {
			return "", fmt.Errorf("option set cannot be a pointer")
		}
		literals := make([]string, len(t.OptionCases))
		seen := map[string]bool{}
		for i, c := range t.OptionCases {
			n := normalize(c, true)
			if !identifier(n) || seen[n] {
				return "", fmt.Errorf("empty or colliding option literal %q", c)
			}
			seen[n] = true
			literals[i] = "'" + n + "'"
		}
		name = "readonly (" + strings.Join(literals, " | ") + ")[]"
	}
	if name != "void" && (t.Nullability == Nullable || ((pointer || t.Reference) && t.Nullability == Unspecified)) {
		name += " | null"
	}
	return name, nil
}

func identifier(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range []byte(s) {
		if !(upper(c) || lower(c) || c == '_' || (i > 0 && digit(c))) {
			return false
		}
	}
	return true
}

// Surface is a checked bijection of qualified names, not unqualified names.
// The identity includes method kind, so +foo and -foo are different originals,
// while the export key intentionally does not: Adamic has one member namespace.
type Surface struct {
	Forward map[string]Output
	Reverse map[string]string
}

// A parent's imported name joins its original one, so a protocol's members stay apart from
// those of a class sharing its Objective-C name (NSAccessibilityElement and its protocol).
func Identity(d Declaration) string {
	parent := d.Parent
	if d.ParentSwiftName != "" {
		parent += "/" + d.ParentSwiftName
	}
	kind := string(d.Kind)
	if d.ClassProperty {
		kind = "class-" + kind
	}
	return kind + ":" + d.Framework + ":" + parent + ":" + d.Name
}

// A callable's key carries its argument shape, so overloads that a call tells apart by its
// arguments (initWithLabel:itemSearchDelegate: and initWithRotorType:itemSearchDelegate:, both
// constructors) are separate exports, while two of one shape still collide.
func ExportKey(d Declaration, o Output) string {
	return memberKey(d, o) + Shape(o.Arguments)
}

// memberKey names a member's slot without its shape: static and instance members are apart, as
// in a class, and only callables may share a slot, as overloads.
func memberKey(d Declaration, o Output) string {
	scope := ""
	if !isType(d.Kind) && exportContext(d) != "" {
		scope = typeName(exportContext(d), d.Framework) + "."
		if d.Kind == ClassMethod || d.ClassProperty {
			scope += "static "
		}
	}
	return o.Module + "#" + scope + o.Name
}

// Shape is a callable's argument shape: positional types, then the options object's sorted fields.
func Shape(layout *Layout) string {
	if layout == nil {
		return ""
	}
	parts := []string{}
	for _, argument := range layout.Positional {
		parts = append(parts, argument.Type)
	}
	options := []string{}
	for _, argument := range layout.Options {
		options = append(options, argument.Name+": "+argument.Type)
	}
	sort.Strings(options)
	if len(options) > 0 {
		parts = append(parts, "{ "+strings.Join(options, "; ")+" }")
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
func explicitName(d Declaration) string {
	if d.SwiftName != "" {
		return d.SwiftName
	}
	return d.APIName
}
func exportContext(d Declaration) string {
	name := strings.Split(explicitName(d), "(")[0]
	if dot := strings.LastIndexByte(name, '.'); dot >= 0 {
		return name[:dot]
	}
	return parentName(d)
}
func Build(declarations []Declaration) (Surface, error) {
	surface := Surface{Forward: map[string]Output{}, Reverse: map[string]string{}}
	// Each slot holds one value member, or callables only, whose shapes then tell them apart.
	slots := map[string]string{}
	// Every problem is reported at once, sorted, so a whole SDK's collisions show in one run and
	// the report never depends on encounter order.
	problems := []string{}
	for _, d := range declarations {
		if d.AccessorProperty != "" {
			return Surface{}, fmt.Errorf("describe accessor %s through its property, with Getter and Setter", d.Name)
		}
		o, err := Name(d)
		if err != nil {
			return Surface{}, err
		}
		original, key := Identity(d), ExportKey(d, o)
		if _, exists := surface.Forward[original]; exists {
			problems = append(problems, "duplicate original "+original)
			continue
		}
		if other, exists := surface.Reverse[key]; exists {
			problems = append(problems, fmt.Sprintf("collision at %s between %s", key, pair(other, original)))
			continue
		}
		slot := memberKey(d, o)
		if other, exists := slots[slot]; exists && (o.Arguments == nil || surface.Forward[other].Arguments == nil) {
			problems = append(problems, fmt.Sprintf("collision at %s between %s", slot, pair(other, original)))
			continue
		}
		slots[slot] = original
		surface.Forward[original] = o
		surface.Reverse[key] = original
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		return Surface{}, fmt.Errorf("%d naming problems:\n%s", len(problems), strings.Join(problems, "\n"))
	}
	return surface, nil
}

// pair names two colliding originals in sorted order.
func pair(a, b string) string {
	if b < a {
		a, b = b, a
	}
	return a + " and " + b
}
