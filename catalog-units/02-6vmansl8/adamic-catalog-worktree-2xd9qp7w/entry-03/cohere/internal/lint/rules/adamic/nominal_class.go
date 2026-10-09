package adamic

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking/flow"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The rule's messages, whose wording lives in `policy/messages/nominal-class.json`.
var (
	nominalClassNotAnInstanceText       = policy.MessageOf("adamic/nominal-class", "notAnInstance")
	nominalClassTypeArgumentsDifferText = policy.MessageOf("adamic/nominal-class", "typeArgumentsDiffer")
)

/*
 * NominalClass reports a value accepted as a class type that is not an instance of that class, or is one
 * with different type arguments (#drbrp8c).
 *
 *     invalid: const shelter: AnimalShelter = new DogShelter();             another class, same shape
 *     invalid: const shelter: AnimalShelter = { admit: (dog: Dog) => '' };  a plain object
 *     invalid: const box: Box<Animal> = dogBox;                             Box<Dog>
 *     valid:   const shelter: Shelter = new DogShelter();                   Shelter is an interface
 *     valid:   const box: Box<Dog> = dogBox;
 *
 * # The hole
 *
 * tsc relates classes structurally, and when the target member is a method it compares parameters in
 * both directions. So `DogShelter`, whose `admit` takes only a Dog, is accepted as `AnimalShelter`, whose
 * `admit` takes any Animal, and `shelter.admit(cat)` is `dog.bark is not a function` (probes h01, h18).
 * `Box<Dog>` passes as `Box<Animal>` the same way, even with a private field (adamic's generic.ts), and
 * `put(cat)` then fills the dog box. Adamic 0.1 makes classes nominal and invariant: a class type accepts
 * an instance of that class, with identical type arguments, and nothing else. An instance can still be
 * passed as an interface it satisfies, which tsc checks strictly when the interface's members are written
 * property-style (probe h02), and method-signature-style sees to that.
 *
 * # How
 *
 * At every pair the file's flow.Walker relates, when the target is a class instance type, the source must be an
 * instance of that class or of one derived from it, and every type argument of the class must be identical.
 * For a derived class that means the arguments it gives the class through its heritage (#ncheh9w): a
 * `class DogBox extends Box<Dog>` is a Box<Dog>, and a `Crate<Dog>` of `class Crate<T> extends Box<T>` is one
 * too, so either seen as a Box<Animal> takes a Cat into the dog's box, as Box<Dog> itself does. The bases are
 * mapped the way TypeScript maps a reference's own (resolveObjectTypeMembers): the declared class's type
 * parameters and `this`, to the reference's arguments, through instantiateType.
 *
 * The rule owns a class instance whole, so it does not descend into one, and invariant-mutable does not
 * either: one hole, one finding. A subclass is accepted, which is right for the value and not for a mutable
 * slot holding it: `Dog[]` seen as `Animal[]` is invariant-mutable's, which judges the slot and leaves a pair
 * this rule reports (nominalReports).
 *
 * # No fix
 *
 * The repair is a different type or a different value, and only the author can pick it.
 */
var NominalClass = rule.Rule{
	Name:             "adamic/nominal-class",
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.SourceFile == nil {
			return nil
		}
		typeChecker := ctx.TypeChecker
		walker := flow.WalkerFor(ctx)
		// The judge and the message it picks are made once per file and set at each site, not a closure per
		// site (#m6tyg79).
		var handle policy.MessageHandle
		judge := func(pair flow.Pair) (bool, bool) {
			if !isClassInstance(pair.Target) {
				return false, true
			}
			verdict := nominalVerdict(typeChecker, pair.Source, pair.Target)
			switch verdict {
			case nominalAccepted:
				return false, false
			case nominalUndecided:
				// A union or intersection source: the walk pairs its members one by one.
				return false, true
			case nominalTypeArgumentsDiffer:
				handle = nominalClassTypeArgumentsDifferText
			default:
				handle = nominalClassNotAnInstanceText
			}
			return true, false
		}
		return walker.Listeners(func(site flow.Site) {
			if site.Spread || site.Method {
				// A spread into a literal and a literal's method are invariant-mutable's; see flow.Site.
				return
			}
			handle = policy.MessageHandle{}
			found, wrong := walker.Walk(site, judge)
			if !wrong {
				return
			}
			ctx.ReportNode(site.Node, rule.Message{
				Id: handle.Id,
				Description: handle.Render(map[string]string{
					"slot":   slotText(ctx.SourceFile, site.Node, found.Path),
					"source": typeChecker.TypeToString(found.Source),
					"target": typeChecker.TypeToString(found.Target),
				}),
			})
		})
	},
}

type nominal int

const (
	nominalAccepted nominal = iota
	nominalUndecided
	nominalNotAnInstance
	nominalTypeArgumentsDiffer
)

// nominalVerdict judges a source against a class instance type.
func nominalVerdict(typeChecker *checker.Checker, source *checker.Type, target *checker.Type) nominal {
	flags := source.Flags()
	switch {
	case flags&(checker.TypeFlagsAny|checker.TypeFlagsNever) != 0:
		// `any` is no-explicit-any's and the no-unsafe rules', and `never` holds no value at all.
		return nominalAccepted
	case flags&checker.TypeFlagsUnion != 0:
		return nominalUndecided
	case flags&checker.TypeFlagsIntersection != 0:
		// One member that is an instance makes the intersection one. Failing that, a member that is an instance
		// with other type arguments says so, as a `DogBox & { tag: string }` seen as a Box<Animal> is one.
		verdict := nominalNotAnInstance
		for _, member := range source.Types() {
			switch nominalVerdict(typeChecker, member, target) {
			case nominalAccepted:
				return nominalAccepted
			case nominalTypeArgumentsDiffer:
				verdict = nominalTypeArgumentsDiffer
			}
		}
		return verdict
	case flags&checker.TypeFlagsTypeParameter != 0:
		// `this` inside a class, or a parameter constrained by one: judged by its constraint.
		constraint := checker.Checker_getBaseConstraintOfType(typeChecker, source)
		if constraint == nil || constraint == source {
			return nominalNotAnInstance
		}
		return nominalVerdict(typeChecker, constraint, target)
	}
	if !isClassInstance(source) {
		return nominalNotAnInstance
	}
	targetClass := classSymbol(target)
	if classSymbol(source) == targetClass {
		return typeArgumentsVerdict(typeChecker, source, target)
	}
	if verdict, found := inheritedVerdict(typeChecker, source, target, targetClass, map[*ast.Symbol]bool{}); found {
		return verdict
	}
	return nominalNotAnInstance
}

// nominalReports is a pair this rule reports: a source that is no instance of the target's class, or one with other
// type arguments.
func nominalReports(typeChecker *checker.Checker, source *checker.Type, target *checker.Type) bool {
	verdict := nominalVerdict(typeChecker, source, target)
	return verdict == nominalNotAnInstance || verdict == nominalTypeArgumentsDiffer
}

// typeArgumentsVerdict judges an instance of the target's own class: every type argument identical.
func typeArgumentsVerdict(typeChecker *checker.Checker, source *checker.Type, target *checker.Type) nominal {
	sourceArguments := checker.Checker_getTypeArguments(typeChecker, source)
	targetArguments := checker.Checker_getTypeArguments(typeChecker, target)
	if len(sourceArguments) != len(targetArguments) {
		return nominalTypeArgumentsDiffer
	}
	for index := range sourceArguments {
		if !checker.Checker_isTypeIdenticalTo(typeChecker, sourceArguments[index], targetArguments[index]) {
			return nominalTypeArgumentsDiffer
		}
	}
	return nominalAccepted
}

// inheritedVerdict walks an instance's bases, with its own type arguments mapped into them, for the target's class,
// and judges the base it reaches as an instance of that class. found is false when no base is the class.
func inheritedVerdict(typeChecker *checker.Checker, instance *checker.Type, target *checker.Type, targetClass *ast.Symbol, seen map[*ast.Symbol]bool) (verdict nominal, found bool) {
	bases, mapped := instantiatedBases(typeChecker, instance)
	for _, base := range bases {
		baseClass := classSymbol(base)
		if baseClass == nil || seen[baseClass] {
			continue
		}
		if baseClass == targetClass {
			if !mapped {
				// Arguments that could not be mapped are not compared, rather than compared in the derived class's
				// own parameters, which would differ from any argument.
				return nominalAccepted, true
			}
			return typeArgumentsVerdict(typeChecker, base, target), true
		}
		seen[baseClass] = true
		if verdict, found := inheritedVerdict(typeChecker, base, target, targetClass, seen); found {
			return verdict, true
		}
	}
	return nominalNotAnInstance, false
}

// instantiatedBases is a class instance's base types with the instance's type arguments in place of the class's
// type parameters, as TypeScript resolves a reference's inherited members: the declared class's parameters and
// its `this`, mapped to the reference's arguments padded with the reference itself as `this`. A class that is not
// a reference to a generic one, or the generic declaration itself, has its bases in its own terms already. mapped
// is false when the arguments do not line up with the parameters, which TypeScript itself never produces.
func instantiatedBases(typeChecker *checker.Checker, instance *checker.Type) (bases []*checker.Type, mapped bool) {
	declared := declaredClassType(typeChecker, instance)
	if declared == nil || declared.ObjectFlags()&(checker.ObjectFlagsClass|checker.ObjectFlagsInterface) == 0 {
		return nil, true
	}
	bases = checker.Checker_getBaseTypes(typeChecker, declared)
	if declared == instance || len(bases) == 0 {
		return bases, true
	}
	parameters := declared.AsInterfaceType().TypeParameters()
	if thisType := declared.AsInterfaceType().ThisType(); thisType != nil {
		parameters = append(slices.Clip(parameters), thisType)
	}
	arguments := checker.Checker_getTypeArguments(typeChecker, instance)
	if len(arguments) == len(parameters)-1 {
		arguments = append(slices.Clip(arguments), instance)
	}
	if len(arguments) != len(parameters) {
		return bases, false
	}
	mapper := checker.NewTypeMapper(parameters, arguments)
	instantiated := make([]*checker.Type, len(bases))
	for index, base := range bases {
		instantiated[index] = checker.Checker_instantiateType(typeChecker, base, mapper)
	}
	return instantiated, true
}

// classSymbol is the class an instance type is an instance of.
func classSymbol(t *checker.Type) *ast.Symbol {
	if t.ObjectFlags()&checker.ObjectFlagsReference != 0 && t.Target() != nil {
		return t.Target().Symbol()
	}
	return t.Symbol()
}

// declaredClassType is the class's own declared type, whose base types the checker resolves.
func declaredClassType(typeChecker *checker.Checker, instance *checker.Type) *checker.Type {
	if instance.ObjectFlags()&checker.ObjectFlagsReference != 0 && instance.Target() != nil {
		return instance.Target()
	}
	return instance
}

// derivesFrom walks a class's base types for the target class.
func derivesFrom(typeChecker *checker.Checker, declared *checker.Type, targetClass *ast.Symbol, seen map[*ast.Symbol]bool) bool {
	if declared == nil || declared.ObjectFlags()&(checker.ObjectFlagsClass|checker.ObjectFlagsInterface) == 0 {
		return false
	}
	for _, base := range checker.Checker_getBaseTypes(typeChecker, declared) {
		baseClass := classSymbol(base)
		if baseClass == nil || seen[baseClass] {
			continue
		}
		if baseClass == targetClass {
			return true
		}
		seen[baseClass] = true
		if derivesFrom(typeChecker, declaredClassType(typeChecker, base), targetClass, seen) {
			return true
		}
	}
	return false
}
