package generate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/system-inc/adamic/internal/apple/naming"
)

// The cycle finder's leaves (docs/apple.md, "Delegates"; ruled October 7). The finder follows what a
// value can reach to refuse a cycle reference counting can't free, and through Apple's declared
// properties nearly every class reaches nearly every other. A leaf is a class it doesn't walk
// through: one whose instances, and the instances of every subclass the SDK declares
// (NSMutableString behind an NSString), hold strong references only to other leaves. Its value is
// all an object of it reaches.
//
// It's derived, never listed by hand. A class's strong references are its declared instance
// properties, its own, its ancestors' and its categories', and its descendants', except those
// declared weak, assign or unsafe_unretained, and the few in documentedUnretained, which the
// headers can't say. What a reference reaches is read from its type: a class, a leaf or not; a
// collection of type arguments (NSArray<NSString *>), which reaches only its arguments when the
// collection's own references are its type parameters or leaves; nothing, for a number, a struct or
// a C string; anything, for id, a block, a Core Foundation object, or a class outside the SDK's
// headers read. Leaves are the greatest set closed under that: assume every class is one, take away
// each that reaches something that isn't, until none changes.
//
// The table lists every leaf with each reference it holds, so a new SDK that adds a reference path
// changes the table, and internal/apple's test that holds the checked-in table to the SDK fails.

// documentedUnretained are references Apple's documentation says an object doesn't keep, where the
// header can't say so: a readonly property with no ownership written, or a parameter only read. A
// method's entry is by selector, for every class that has it.
var documentedUnretained = map[string]string{
	"NSValue.nonretainedObjectValue": "NSValue.h: valueWithNonretainedObject: keeps the object without retaining it, and nonretainedObjectValue returns it",
	"initWithCoder:":                 "NSCoding: an object initialized from a coder decodes what it holds from it and never keeps the coder",
	"NSKeyValueObserverRegistration": "NSKeyValueObserving.h: key-value observing doesn't retain an observer, and removing one keeps nothing",
	"valueWithNonretainedObject:":    "NSValue.h: the value keeps the object without retaining it",
	"filterUsingPredicate:":          "NSPredicate.h: a collection filtered by a predicate keeps the elements that match, not the predicate",
	"setValue:forKey:":               "NSKeyValueCoding.h: a collection's setValue:forKey: sets the key on each of its elements and keeps nothing itself; on any other class it sets a property, whose own type is counted",
	"locale":                         "NSLocale.h: a locale handed to parse or format (initWithString:locale:, descriptionWithLocale:) is read, never kept",
	"URLFromPasteboard:":             "NSPasteboard.h: reads a URL off the pasteboard; the URL keeps nothing of it",
}

// heldKind is what one reference's type reaches.
type heldKind int

const (
	heldNothing   heldKind = iota + 1 // a number, a struct, a C string: no object
	heldAnything                      // id, a block, a Core Foundation object, an unknown class
	heldParameter                     // the class's own type parameter (NSArray's ObjectType)
	heldClass                         // an Objective-C class, with type arguments
)

type held struct {
	kind      heldKind
	class     string
	arguments []held
	text      string
}

// strongReference is one reference a class's instances may hold strongly: a property, or what a
// method is handed.
type strongReference struct {
	// through is where it comes from: NSURL.baseURL, or -[NSData initWithContentsOfURL:].
	through string
	held    held
}

// heldType reads what a property's type reaches.
func (g *generator) heldType(written writtenType, parameters map[string]bool) held {
	text := written.Qual
	if written.Desugared != "" && !strings.Contains(written.Qual, "<") && !parameters[cleanHeld(written.Qual)] {
		text = written.Desugared
	}
	return g.heldText(cleanHeld(withoutAttributes(text)), parameters)
}

func cleanHeld(text string) string {
	for _, word := range []string{"_Nullable", "_Nonnull", "_Null_unspecified", "__kindof", "__strong", "__unsafe_unretained", "__autoreleasing", "const"} {
		text = strings.ReplaceAll(text, word, " ")
	}
	return strings.Join(strings.Fields(text), " ")
}

func (g *generator) heldText(text string, parameters map[string]bool) held {
	result := held{text: text}
	switch {
	case strings.Contains(text, "^"):
		result.kind = heldAnything
	case parameters[text]:
		result.kind = heldParameter
	case text == "id" || strings.HasPrefix(text, "id<") || strings.HasPrefix(text, "id <"):
		result.kind = heldAnything
	case !strings.HasSuffix(text, "*"):
		// A number, a struct, a selector, a class object (which lives as long as the process).
		result.kind = heldNothing
	case parameters[strings.TrimSpace(strings.TrimSuffix(text, "*"))]:
		// A C array of the type parameter (initWithObjects:count:): its elements.
		result.kind = heldParameter
	default:
		base := strings.TrimSpace(strings.TrimSuffix(text, "*"))
		name, arguments, generic := strings.Cut(base, "<")
		name = strings.TrimSpace(name)
		if close := strings.LastIndex(base, ">"); generic && close >= 0 && strings.Contains(base[close:], "*") || !generic && strings.Contains(base, "*") {
			// A pointer to a pointer: an out parameter's shape, whatever it points at.
			result.kind = heldAnything
			return result
		}
		if def := g.types[name]; def != nil && def.declaration.Kind == naming.Class {
			result.kind, result.class = heldClass, name
			if generic {
				for _, argument := range splitTopLevel(strings.TrimSuffix(strings.TrimSpace(arguments), ">")) {
					result.arguments = append(result.arguments, g.heldText(cleanHeld(argument), parameters))
				}
			}
			return result
		}
		switch base {
		case "void", "char", "unsigned char", "signed char", "unichar", "uint8_t", "int8_t", "int", "unsigned int", "short", "unsigned short", "long", "unsigned long", "float", "double", "NSInteger", "NSUInteger", "CGFloat":
			// A C pointer to bytes or numbers: memory, not an object.
			result.kind = heldNothing
		default:
			result.kind = heldAnything
		}
	}
	return result
}

// leaves derives the leaf table.
func (g *generator) leaves() []byte {
	classes := []string{}
	for _, name := range sortedKeys(g.types) {
		if g.types[name].declaration.Kind == naming.Class && g.types[name].node.Name == name {
			classes = append(classes, name)
		}
	}
	descendants := map[string][]string{}
	for _, name := range classes {
		for parent := g.types[name].node.Super.Name; parent != ""; {
			descendants[parent] = append(descendants[parent], name)
			def := g.types[parent]
			if def == nil {
				break
			}
			parent = def.node.Super.Name
		}
	}
	// Each class's own strong references: its properties and its categories'.
	own := map[string][]strongReference{}
	exceptions := map[string]bool{}
	for _, name := range classes {
		def := g.types[name]
		parameters := map[string]bool{}
		for _, child := range children(def.node, "ObjCTypeParamDecl") {
			parameters[child.Name] = true
		}
		properties := children(def.node, "ObjCPropertyDecl")
		for _, category := range g.nodes {
			// A category on the root class is an informal protocol, as emitClass reads one: what a
			// delegate may implement, declared on NSObject for the compiler, not what NSObject holds.
			if category.Kind == "ObjCCategoryDecl" && category.Interface.Name == name && def.node.Super.Name != "" {
				properties = append(properties, children(category, "ObjCPropertyDecl")...)
			}
		}
		for _, property := range properties {
			if property.ClassProperty || property.Unowned {
				continue
			}
			if reason, documented := documentedUnretained[name+"."+property.Name]; documented {
				exceptions[name+"."+property.Name+": "+reason] = true
				continue
			}
			reference := strongReference{through: name + "." + property.Name, held: g.heldType(property.Type, parameters)}
			if reference.held.kind != heldNothing {
				own[name] = append(own[name], reference)
			}
		}
		// What a method is handed, an instance may keep (setObject:forKey:, addObserver:): every
		// object and escaping block a bound method of the class, or of a category, takes, since a
		// method no binding sends is one a program can't hand anything. A pointer to a pointer is an
		// out parameter, and a block marked NS_NOESCAPE can't outlive the call. The root class's own
		// messages (isEqual:, performSelector:withObject:) are every object's and keep nothing.
		if def.node.Super.Name == "" {
			continue
		}
		methods := children(def.node, "ObjCMethodDecl")
		for _, category := range g.nodes {
			if category.Kind != "ObjCCategoryDecl" || category.Interface.Name != name {
				continue
			}
			if reason, documented := documentedUnretained[category.Name]; documented {
				exceptions["category "+category.Name+": "+reason] = true
				continue
			}
			methods = append(methods, children(category, "ObjCMethodDecl")...)
		}
		for _, method := range methods {
			if method.Implicit || !g.bound[methodOriginal(name, method.Name, method.Instance)] {
				continue
			}
			if reason, documented := documentedUnretained[method.Name]; documented {
				exceptions[method.Name+": "+reason] = true
				continue
			}
			// What an initializer or a class's factory is handed, the object it makes may keep
			// (scheduledTimerWithTimeInterval:target:..., initWithTarget:selector:object:), and so may
			// a mutator, an instance method that gives nothing back (setObject:forKey:,
			// addObserver:selector:name:object:) or one Cocoa's naming conventions name one
			// (setDataProvider:forTypes: gives a BOOL). An escaping block is kept by whatever takes it.
			// Any other instance method reads its arguments (descriptionWithLocale:), and keeps what it
			// keeps in what it gives back.
			reads := method.Instance && cleanType(method.Result.Qual) != "void" && !mutatorName(method.Name) && !strings.HasPrefix(method.Name, "init")
			for _, parameter := range children(method, "ParmVarDecl") {
				if has(parameter, "NoEscapeAttr") {
					continue
				}
				if reason, documented := documentedUnretained[parameter.Name]; documented {
					exceptions["parameter "+parameter.Name+": "+reason] = true
					continue
				}
				if reads && !strings.Contains(parameter.Type.Qual, "^") && !strings.Contains(parameter.Type.Desugared, "^") {
					continue
				}
				reference := strongReference{through: methodOriginal(name, method.Name, method.Instance), held: g.heldType(parameter.Type, parameters)}
				if reference.held.kind == heldNothing || reference.held.kind == heldAnything && strings.Count(strings.SplitN(reference.held.text, "<", 2)[0]+tail(reference.held.text), "*") > 1 {
					continue
				}
				own[name] = append(own[name], reference)
			}
		}
	}
	// What an instance of a class may hold: its ancestors' references, its own, and every
	// descendant's, since an object of the class may be one of theirs.
	holds := map[string][]strongReference{}
	for _, name := range classes {
		references := append([]strongReference{}, own[name]...)
		for parent := g.types[name].node.Super.Name; parent != ""; {
			references = append(references, own[parent]...)
			def := g.types[parent]
			if def == nil {
				break
			}
			parent = def.node.Super.Name
		}
		for _, descendant := range descendants[name] {
			references = append(references, own[descendant]...)
		}
		holds[name] = references
	}
	generic := func(name string) bool {
		return len(children(g.types[name].node, "ObjCTypeParamDecl")) > 0
	}
	leaf, container := map[string]bool{}, map[string]bool{}
	for _, name := range classes {
		// A root class (NSObject, NSProxy) is every class's ancestor: an object of it may be anything.
		// A class whose interface the headers read declares nothing of (NSManagedObjectContext, which
		// CoreData defines and AppKit only extends) holds what they don't show. A generic class is a collection, whatever its methods show: a collection is never a
		// leaf, though one may reach only its type arguments.
		declared := len(children(g.types[name].node, "ObjCMethodDecl"))+len(children(g.types[name].node, "ObjCPropertyDecl")) > 0
		known := g.types[name].node.Super.Name != "" && g.selected(g.types[name].node.Framework) && declared
		leaf[name] = known && !generic(name)
		container[name] = known && generic(name)
	}
	var reachesLeaves func(reference held, parameters bool) bool
	reachesLeaves = func(reference held, parameters bool) bool {
		switch reference.kind {
		case heldNothing:
			return true
		case heldParameter:
			return parameters
		case heldClass:
			if leaf[reference.class] {
				return true
			}
			if !container[reference.class] || len(reference.arguments) == 0 {
				return false
			}
			for _, argument := range reference.arguments {
				if !reachesLeaves(argument, parameters) {
					return false
				}
			}
			return true
		}
		return false
	}
	// A leaf's references reach only leaves; a collection's may reach its type parameters too.
	for changed := true; changed; {
		changed = false
		for _, name := range classes {
			for _, table := range []struct {
				members    map[string]bool
				parameters bool
			}{{leaf, false}, {container, true}} {
				if !table.members[name] {
					continue
				}
				for _, reference := range holds[name] {
					if !reachesLeaves(reference.held, table.parameters) {
						table.members[name], changed = false, true
						break
					}
				}
			}
		}
	}

	var text strings.Builder
	text.WriteString("# The classes Adamic's cycle finder doesn't walk through: each holds strong references only to\n")
	text.WriteString("# other leaves, and so does every subclass the headers declare (docs/apple.md, \"Delegates\").\n")
	text.WriteString("# Generated by internal/apple/generate (leaves.go) from the headers. Do not edit.\n")
	text.WriteString("# A line under a class is a strong reference an instance of it, or of a subclass, may hold:\n")
	text.WriteString("# the property or the method it comes through, its type, and why it reaches only leaves.\n")
	for _, exception := range sortedKeys(exceptions) {
		text.WriteString("\nunretained " + exception + "\n")
	}
	describe := func(reference held) string {
		if reference.kind == heldClass && leaf[reference.class] {
			return "a leaf"
		}
		if reference.kind == heldClass {
			return "a collection of leaves"
		}
		if reference.kind == heldParameter {
			return "its type parameter"
		}
		return "no object"
	}
	for _, kind := range []string{"leaf", "container"} {
		for _, name := range classes {
			if kind == "leaf" && !leaf[name] || kind == "container" && (leaf[name] || !container[name]) {
				continue
			}
			fmt.Fprintf(&text, "\n%s %s\n", kind, name)
			lines := []string{}
			for _, reference := range holds[name] {
				lines = append(lines, fmt.Sprintf("\t%s %s: %s", reference.through, reference.held.text, describe(reference.held)))
			}
			sort.Strings(lines)
			for _, line := range lines {
				text.WriteString(line + "\n")
			}
		}
	}
	return []byte(text.String())
}

// tail is what follows a type's generic arguments: NSArray<NSString *> ** has " **".
func tail(text string) string {
	if close := strings.LastIndex(text, ">"); close >= 0 {
		return text[close+1:]
	}
	return ""
}

// mutatorName reports whether Cocoa's naming conventions make a selector a mutator: it sets, adds,
// inserts, appends, registers or schedules what it's handed.
func mutatorName(selector string) bool {
	for _, verb := range []string{"set", "add", "insert", "append", "register", "schedule", "replace", "push", "enqueue"} {
		if strings.HasPrefix(selector, verb) && len(selector) > len(verb) && (selector[len(verb)] < 'a' || selector[len(verb)] > 'z') {
			return true
		}
	}
	return false
}

// selected reports whether a framework is one of those generated.
func (g *generator) selected(framework string) bool {
	for _, candidate := range g.configuration.Frameworks {
		if candidate.Name == framework {
			return true
		}
	}
	return false
}
