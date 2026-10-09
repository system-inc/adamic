package adamic

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/checking/flow"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The rule's messages, whose wording lives in `policy/messages/invariant-mutable.json`.
var (
	invariantMutableText     = policy.MessageOf("adamic/invariant-mutable", "mutableWidening")
	readonlyMadeWritableText = policy.MessageOf("adamic/invariant-mutable", "readonlyMadeWritable")
	methodParameterText      = policy.MessageOf("adamic/invariant-mutable", "methodParameterNarrowed")
)

/*
 * InvariantMutable reports a value seen through a wider type at a place that type lets you write (#drbrp8c).
 *
 *     invalid: const animals: Animal[] = dogs;             animals.push(cat) puts a cat in dogs
 *     invalid: const pen: { pet: Animal } = kennel;         pen.pet = cat does the same
 *     invalid: const animals: Map<string, Animal> = dogs;   animals.set('Tom', cat)
 *     valid:   const animals: readonly Animal[] = dogs;     nothing can be written through it
 *     valid:   const pen: { pet: Animal } = { pet: rex };   the literal is held by nobody else
 *
 * # The hole
 *
 * tsc relates a mutable location covariantly: an element of `Dog[]` is accepted where an element of
 * `Animal[]` is wanted, because reading one is safe. Writing one is not, and tsc lets the wider name
 * write. Each shape above is accepted by tsc 6.0.3 under strict and fails on Node with `dog.bark is not
 * a function` (probes h03, h04, h13, h19, h20 on #drbrp8c). Adamic's rule: a mutable location is
 * invariant and a `readonly` one is covariant, which is what tsc already does for `readonly T[]`.
 *
 * # How
 *
 * The file's flow.Walker finds every place a value goes into a typed slot and pairs the value's parts
 * with the slot's. At every pair the target marks mutable, whatever the wider type can write there must be
 * something the original can read: the target's slot type must be assignable to the source's. tsc already
 * proved the other direction when it accepted the site, so together the slot is invariant. That catches a
 * literal widened in a mutable slot (`{ kind: 'Circle' }` into `{ kind: string }`) and a callback slot
 * narrowed the wrong way, and leaves alone a slot whose two types differ only where writing back is safe, as
 * a DOM event handler's `this` does on every element seen as `Element`.
 *
 * A read-only source slot made writable is the same hole one step removed (#gvzdft9): `readonly pet: Animal`
 * is covariant, so it may hold a Dog, and `{ pet: Animal }` over it writes a Cat there. tsc accepts a read-only
 * property as a writable one, though it refuses `readonly T[]` as `T[]`, so the rule reports the property
 * whatever its two types are; readonlyMadeWritable says where not.
 *
 * A method taking a narrower parameter than the wider type passes is the hole a caller opens (#gvzdft9 shape 5):
 * tsc compares a method's parameters in both directions, so `{ put(dog: Dog) }` passes as `{ put(animal: Animal):
 * void }`, and a Cat passed through the wider type reaches code reading a Dog. The rule asks the walk for methods
 * and judges their parameters one way, a function's being tsc's own under strictFunctionTypes. A generic method is
 * related as tsc relates it (#bbtfx99): one with type parameters of its own is instantiated in the context of the
 * other, so `{ put<T extends Dog>(dog: T) }` as `{ put(animal: Animal) }` is a Dog parameter under an Animal one,
 * and Promise's `then` is the same `then` under two promises. A method's return is paired as a function's is
 * (#y0ejf6a): `all(): Dog[] { return this.dogs }` seen as `all(): Animal[]` hands out the array the object still
 * holds. A method whose every return is a literal or a new container builds what it returns, and the library's
 * methods are taken to (an iterator's `next()`), so those returns' own slots are no one else's.
 *
 * It began as identity, by the checker's identity relation, and the four consumers measured why not: an
 * `HTMLElement` seen as an `Element` reported its event handler slots, which differ only in `this` and are
 * safe to write back. `{ x }` against `{ x; y?: number }` is mutually assignable, and that hole is
 * no-optional-widening's.
 *
 * A class instance is nominal-class's: it requires identical type arguments, which covers every mutable
 * part, so this rule does not descend into one. The slot that holds one is this rule's (#5y54eyj): nominal-class
 * rightly accepts a Dog as an Animal, which is the value, and `Dog[]` seen as `Animal[]` is still a cat written
 * into the dogs, a TypeError in Node whether the animals are interfaces or classes. Where nominal-class reports
 * the pair itself, a value that is no instance or one with other type arguments, this rule leaves it, so the
 * two rules never report one hole twice.
 *
 * # No fix
 *
 * Every repair changes a type (add `readonly`, or copy), so none preserves meaning, and the edit engine
 * applies only fixes that do.
 */
var InvariantMutable = rule.Rule{
	Name:             "adamic/invariant-mutable",
	NeedsTypeChecker: true,
	ProgramReads:     rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.SourceFile == nil {
			return nil
		}
		typeChecker := ctx.TypeChecker
		walker := flow.WalkerFor(ctx)
		// Made once per file, not per site: a judge made per site was a closure per site (#m6tyg79).
		judge := func(pair flow.Pair) (bool, bool) {
			// What a class instance is, is nominal-class's, so the walk does not go into one. The slot holding it
			// is still judged here, as any slot is (#5y54eyj).
			descend := !isClassInstance(pair.Target)
			// A method's parameter is written by the caller, so whatever the wider type passes must be something
			// the method reads: the target's parameter type (the pair's Source) assignable to the source's. tsc
			// compares a method's parameters both ways (#gvzdft9 shape 5); a function's it already checks one way
			// under strictFunctionTypes, so only a method's are judged here. The parameter is judged whole, as tsc
			// judges a function's, and the walk goes no further into it: below a callback parameter is the value it is
			// called with, and Promise's `then` read every promise's value as a slot written through the wider type,
			// about a hundred findings on fresh values, an async function's literal or `Promise.resolve([])` (#bbtfx99).
			if pair.MethodParameter {
				return !walker.IsAssignableToParameter(pair.Source, pair.Target, pair.Defaulted), false
			}
			if !pair.Mutable {
				return false, descend
			}
			// Writing through the wider type stores a target value where the original reads a source
			// value, so every target value must be a source value too.
			wrong := readonlyMadeWritable(ctx.Program, pair) || !walker.IsAssignable(pair.Target, pair.Source)
			if wrong && !descend && nominalReports(typeChecker, pair.Source, pair.Target) {
				// Not an instance of the class, or one with other type arguments: nominal-class's finding, on
				// this pair, and one hole is one finding.
				return false, false
			}
			return wrong, descend
		}
		return walker.Listeners(func(site flow.Site) {
			// Object intersections are this rule's to pair, and no other's (#b9a0wgy).
			found, wrong := walker.WalkWith(site, judge, flow.WalkOptions{ObjectIntersections: true, Methods: true})
			if !wrong {
				return
			}
			handle, id := invariantMutableText, "mutableWidening"
			switch {
			case found.MethodParameter:
				handle, id = methodParameterText, "methodParameterNarrowed"
			case readonlyMadeWritable(ctx.Program, found) && walker.IsAssignable(found.Target, found.Source):
				handle, id = readonlyMadeWritableText, "readonlyMadeWritable"
			}
			ctx.ReportNode(site.Node, rule.Message{
				Id: id,
				Description: handle.Render(map[string]string{
					"source":     type_checking.StableTypeText(typeChecker, site.Source),
					"target":     type_checking.StableTypeText(typeChecker, site.Target),
					"slot":       slotText(ctx.SourceFile, site.Node, found.Path),
					"sourcePart": type_checking.StableTypeText(typeChecker, found.Source),
					"targetPart": type_checking.StableTypeText(typeChecker, found.Target),
				}),
			})
		})
	},
}

/*
 * readonlyMadeWritable is a read-only source slot under a writable target slot (#gvzdft9): the value behind
 * `readonly pet: Animal` may be a Dog, since a read-only slot is covariant, and the writable view stores any
 * Animal there. Not when the slot's type is a unit type (a literal, `null`), which holds nothing narrower, and
 * not when the property is declared in TypeScript's default library: `readonly` there is a WebIDL readonly
 * attribute or an accessor with no setter, and a write through any view throws in strict code before anything
 * is stored (Node: `Cannot set property readable of #<TransformStream> which has only a getter`), as a class's
 * `prototype` does (#b9a0wgy). Measured on the consumers, TransformStream's `readable` seen through
 * pipeThrough's ReadableWritablePair and SVG's `className` seen as an Element's were 17 of a 24-finding sample.
 */
func readonlyMadeWritable(program rule.Program, pair flow.Pair) bool {
	if !pair.SourceReadonly || pair.Source.Flags()&checker.TypeFlagsUnit != 0 || pair.SourceProperty == nil {
		return false
	}
	for _, declaration := range pair.SourceProperty.Declarations {
		if file := ast.GetSourceFileOfNode(declaration); file != nil && program != nil && type_checking.IsSourceFileDefaultLibrary(program, file) {
			return false
		}
	}
	return true
}

// slotText names the mutable slot as a reader would reach it from the expression: `dogs[]`,
// `kennel.pet`, `dogs<Map value>`. A long or multi-line expression is named `value` instead.
func slotText(sourceFile *ast.SourceFile, node *ast.Node, path []flow.Step) string {
	textRange := rule.TokenRange(sourceFile, node)
	expression := sourceFile.Text()[textRange.Pos():textRange.End()]
	if len(expression) > 40 || strings.ContainsAny(expression, "\n\r") {
		expression = "value"
	}
	return expression + flow.PathText(path)
}

// isClassInstance is the instance side of a class, generic or not: `Box<Dog>`, `Shelter`. The class's
// own constructor type (`typeof Shelter`) carries the class's symbol too, and is not one.
func isClassInstance(t *checker.Type) bool {
	if t == nil || t.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if t.ObjectFlags()&checker.ObjectFlagsClass != 0 {
		return true
	}
	return t.ObjectFlags()&checker.ObjectFlagsReference != 0 && t.Target() != nil &&
		t.Target().ObjectFlags()&checker.ObjectFlagsClass != 0
}
