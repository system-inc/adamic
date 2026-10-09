package adamic

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// shelters is probe h01's two classes: the same shape, and a method that takes only a Dog where the other
// takes any Animal.
const shelters = animals + `
class AnimalShelter { admit(animal: Animal): string { return animal.name; } }
class DogShelter { admit(dog: Dog): string { return dog.bark(); } }
class Box<Item> {
	private item: Item;
	constructor(item: Item) { this.item = item; }
	put(item: Item): void { this.item = item; }
	get(): Item { return this.item; }
}
`

// heritage is shelters with Box's subclasses (#ncheh9w): one fixing its argument, one passing its own through,
// one fixing that one's, and one passing an array of its own. heritage-probes.tgz on the task runs the first two
// in Node: seen as a Box<Animal>, a Cat put in, and the dog's `bark` is not a function.
const heritage = shelters + `
class DogBox extends Box<Dog> {}
class Crate<Contents> extends Box<Contents> {}
class DogCrate extends Crate<Dog> {}
class Pen<Member> extends Box<Member[]> {}
`

// TestNominalClassFiresOnEveryProvenHole: each case is accepted by tsc 6.0.3 and fails on Node.
func TestNominalClassFiresOnEveryProvenHole(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		id     string
		span   string
	}{
		// h01
		"another class": {shelters + `const shelter: AnimalShelter = new DogShelter();`, "notAnInstance", "new DogShelter()"},
		// h18
		"a plain object": {shelters + `const shelter: AnimalShelter = { admit: (dog: Dog): string => dog.bark() };`, "notAnInstance", "{ admit: (dog: Dog): string => dog.bark() }"},
		// adamic's generic.ts
		"different type arguments":      {shelters + `const dogBox = new Box<Dog>(rex); const animalBox: Box<Animal> = dogBox;`, "typeArgumentsDiffer", "dogBox"},
		"an argument":                   {shelters + `function house(shelter: AnimalShelter): void { shelter.admit({ name: 'Tom' }); } house(new DogShelter());`, "notAnInstance", "new DogShelter()"},
		"a class in a property":         {shelters + `const dogs = { shelter: new DogShelter() }; const all: { readonly shelter: AnimalShelter } = dogs;`, "notAnInstance", "dogs"},
		"an empty class takes anything": {`class Token {} const token: Token = 42;`, "notAnInstance", "42"},
		// Strict on `any` too: a Box<any> seen as a Box<Dog> is the same hole the any opened, reported where it is used.
		"an any type argument": {shelters + `declare const loose: Box<any>; const strict: Box<Dog> = loose;`, "typeArgumentsDiffer", "loose"},
		// heritage-probes.tgz: a subclass gives Box its arguments, and those are what must be identical (#ncheh9w).
		"a subclass fixing another argument":          {heritage + `const animalBox: Box<Animal> = new DogBox(rex);`, "typeArgumentsDiffer", "new DogBox(rex)"},
		"a generic subclass passing another argument": {heritage + `const animalBox: Box<Animal> = new Crate<Dog>(rex);`, "typeArgumentsDiffer", "new Crate<Dog>(rex)"},
		"two levels of subclass":                      {heritage + `const animalBox: Box<Animal> = new DogCrate(rex);`, "typeArgumentsDiffer", "new DogCrate(rex)"},
		"an argument mapped into another type":        {heritage + `const animalBox: Box<Animal[]> = new Pen<Dog>([rex]);`, "typeArgumentsDiffer", "new Pen<Dog>([rex])"},
		// api's drizzle adapters: an intersection whose member is the subclass says the arguments differ.
		"a subclass in an intersection": {heritage + `declare const tagged: DogBox & { readonly tag: string }; const animalBox: Box<Animal> = tagged;`, "typeArgumentsDiffer", "tagged"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, NominalClass, fixture.source)
			rule_testing.ExpectFindings(t, result, fixture.id)
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestNominalClassStaysCleanOnInstancesAndInterfaces: the class's own instances, its subclasses, and an
// instance passed as an interface it satisfies.
func TestNominalClassStaysCleanOnInstancesAndInterfaces(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"its own instance":       shelters + `const shelter: AnimalShelter = new AnimalShelter();`,
		"the same type argument": shelters + `const dogBox = new Box<Dog>(rex); const same: Box<Dog> = dogBox;`,
		// h02's shape, with the interface tsc checks strictly
		"an interface it satisfies": shelters + `interface Shelter { readonly admit: (animal: Animal) => string } const shelter: Shelter = new AnimalShelter();`,
		"a subclass":                `class Base { readonly id = 1; } class Derived extends Base { readonly extra = 2; } const base: Base = new Derived();`,
		"this":                      `class Chain { next(): Chain { return this; } }`,
		"a union of instances":      shelters + `declare const either: AnimalShelter | undefined; const maybe: AnimalShelter | undefined = either;`,
		// The heritage near-misses (#ncheh9w): the argument the subclass gives, and a read-only view, not a class.
		"a subclass fixing the same argument":          heritage + `const dogs: Box<Dog> = new DogBox(rex);`,
		"a generic subclass passing the same argument": heritage + `const dogs: Box<Dog> = new Crate<Dog>(rex);`,
		"two levels with the same argument":            heritage + `const dogs: Box<Dog> = new DogCrate(rex); const crate: Crate<Dog> = new DogCrate(rex);`,
		"an argument mapped the same way":              heritage + `const dogs: Box<Dog[]> = new Pen<Dog>([rex]);`,
		"a read-only view of a subclass":               heritage + `const view: { readonly get: () => Animal } = new DogBox(rex);`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, NominalClass, source))
		})
	}
}

// TestNominalClassStaysCleanOnAdamicsPrograms: Adamic 0.1 compiles all ten, the generic Stack among them.
func TestNominalClassStaysCleanOnAdamicsPrograms(t *testing.T) {
	t.Parallel()
	expectProgramsClean(t, NominalClass)
}
