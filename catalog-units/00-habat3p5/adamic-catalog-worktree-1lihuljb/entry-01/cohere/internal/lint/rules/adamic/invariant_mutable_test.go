package adamic

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// animals is the probes' shared vocabulary: a Dog is an Animal that barks, and a cat written through an
// Animal-typed name is what makes each hole a TypeError.
const animals = `
interface Animal { readonly name: string }
interface Dog extends Animal { readonly bark: () => string }
declare const rex: Dog;
`

// TestInvariantMutableFiresOnEveryProvenHole: each case is a probe tsc 6.0.3 accepts and Node fails on,
// and the finding sits on the value being widened.
func TestInvariantMutableFiresOnEveryProvenHole(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		span   string
	}{
		// h03
		"a mutable property": {animals + `
const kennel: { pet: Dog } = { pet: rex };
const pen: { pet: Animal } = kennel;
pen.pet = { name: 'Tom' };`, "kennel"},
		// h04
		"a map's value": {animals + `
const dogs = new Map<string, Dog>();
const others: Map<string, Animal> = dogs;`, "dogs"},
		// h13
		"an argument": {animals + `
function adopt(all: Animal[]): void { all.push({ name: 'Tom' }); }
const dogs: Dog[] = [rex];
adopt(dogs);`, "dogs"},
		// h19
		"a returned type parameter array": {animals + `
function widen<Item extends Animal>(items: Item[]): Animal[] { return items; }`, "items"},
		// h20
		"an upcast": {animals + `
const dogs: Dog[] = [rex];
(dogs as Animal[]).push({ name: 'Tom' });`, "dogs"},
		"an assignment": {animals + `
const dogs: Dog[] = [rex];
let all: Animal[] = [];
all = dogs;`, "dogs"},
		"a mutable array under a readonly property": {animals + `
const kennel: { readonly dogs: Dog[] } = { dogs: [rex] };
const pen: { readonly dogs: Animal[] } = kennel;`, "kennel"},
		"an array inside an object literal": {animals + `
const dogs: Dog[] = [rex];
const pen: { readonly all: Animal[] } = { all: dogs };`, "dogs"},
		"a literal type in a mutable slot": {`
const circle: { kind: 'Circle' } = { kind: 'Circle' };
const shape: { kind: string } = circle;`, "circle"},
		"a union slot that may be missing": {animals + `
const dogs: Dog[] = [rex];
const maybe: Animal[] | undefined = dogs;`, "dogs"},
		"shared elements in a new array": {animals + `
const kennels: { pet: Dog }[] = [{ pet: rex }];
const pens: { pet: Animal }[] = kennels.map((kennel) => kennel);`, "kennels.map((kennel) => kennel)"},
		// t4 on #drbrp8c: a function can hand back what it holds, so its result is walked like any value.
		"a function's shared result": {animals + `
const kennel: Dog[] = [rex];
const dogsOf = (): Dog[] => kennel;
const animalsOf: () => Animal[] = dogsOf;`, "dogsOf"},
		// #jpdf48x: a literal returned holds what the function does one level down, so that level is still shared.
		"a held array in a returned literal": {animals + `
const kennel: Dog[] = [rex];
const kennelOf = (): { pets: Dog[] } => ({ pets: kennel });
const wide: () => { pets: Animal[] } = kennelOf;`, "kennelOf"},
		// #qq4haam, const-probe.tgz's r1 and r2: a local value built here, which escapes before the return.
		"a local value also cached": {animals + `
const cache: Dog[][] = [];
const kennelOf = (): { pets: Dog[] } => { const pets = [rex]; cache.push(pets); return { pets }; };
const wide: () => { pets: Animal[] } = kennelOf;`, "kennelOf"},
		// A `.map` handing back the elements it is given builds the array, not what is in it.
		"a local map of held elements": {`interface Row { id?: string } declare const rows: { id: string }[]; const make = (): { rows: { id: string }[] } => { const kept = rows.map((row) => row); return { rows: kept }; }; const wide: () => { rows: Row[] } = make;`, "make"},
		"a local value captured": {animals + `
let later: () => Dog[] = () => [];
const kennelOf = (): { pets: Dog[] } => { const pets = [rex]; later = () => pets; return { pets }; };
const wide: () => { pets: Animal[] } = kennelOf;`, "kennelOf"},
		"a module's value, not a local": {animals + `
const pets = [rex];
const kennelOf = (): { pets: Dog[] } => { return { pets }; };
const wide: () => { pets: Animal[] } = kennelOf;`, "kennelOf"},
		// A spread after the last write may be what the literal holds there.
		"a part a later spread may write": {animals + `
declare const holder: { pets: Dog[] };
const kennelOf = (): { pets: Dog[] } => ({ pets: [rex], ...holder });
const wide: () => { pets: Animal[] } = kennelOf;`, "kennelOf"},
		// #jpdf48x, function-probe.tgz's r2: one return of two hands back what the function holds.
		"a function's one held return": {animals + `
const kennel: Dog[] = [rex];
function dogsOf(fresh: boolean): Dog[] { if(fresh) { return []; } return kennel; }
const animalsOf: (fresh: boolean) => Animal[] = dogsOf;`, "dogsOf"},
		"a mutable callback slot": {animals + `
const handler: { onAnimal: (animal: Dog) => void } = { onAnimal: (dog) => { dog.bark(); } };
const wide: { onAnimal: (animal: never) => void } = handler;`, "handler"},
		// #53w68gt: a type parameter target is written through its constraint, and pack.push stores a
		// Pack[number] where narrow holds only Narrow[number].
		"a type parameter target": {animals + `
function admit<Pack extends Animal[], Narrow extends Pack>(narrow: Narrow, extra: Pack[number]): void {
	const pack: Pack = narrow;
	pack.push(extra);
}`, "narrow"},
		"a type parameter target behind a readonly property": {animals + `
function admit<Pack extends { readonly all: Animal[] }, Narrow extends Pack>(narrow: Narrow): Pack {
	return narrow;
}`, "narrow"},
		// #53w68gt: an intersection's array member is a slot like any array.
		"an intersection target": {animals + `
declare const tagged: Dog[] & { tag: string };
const wide: Animal[] & { tag: string } = tagged;`, "tagged"},
		"an intersection source": {animals + `
declare const tagged: Dog[] & { tag: string };
const all: Animal[] = tagged;`, "tagged"},
		// #b9a0wgy: an intersection of objects, as either side; pen.pet = cat reaches kennel.
		"an object intersection target": {animals + `
interface Named { readonly label: string }
const kennel: { pet: Dog } & Named = { pet: rex, label: 'kennel' };
const pen: { pet: Animal } & Named = kennel;`, "kennel"},
		"an object intersection source": {animals + `
interface Named { readonly label: string }
const kennel: { pet: Dog } & Named = { pet: rex, label: 'kennel' };
const pen: { pet: Animal } = kennel;`, "kennel"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, InvariantMutable, fixture.source)
			rule_testing.ExpectFindings(t, result, "mutableWidening")
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestInvariantMutableStaysCleanWhereNothingCanBeWrittenThrough: read-only targets, fresh literals, the
// same type, and a class instance (nominal-class's) are not this rule's holes.
func TestInvariantMutableStaysCleanWhereNothingCanBeWrittenThrough(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		// h14
		"a readonly array": animals + `const dogs: Dog[] = [rex]; const all: readonly Animal[] = dogs;`,
		// h11
		"a fresh literal":       animals + `const pen: { pet: Animal } = { pet: rex };`,
		"a fresh array literal": animals + `const all: Animal[] = [rex, rex];`,
		"a spread copy":         animals + `const dogs: Dog[] = [rex]; const all: Animal[] = [...dogs];`,
		"a readonly property":   animals + `const kennel: { pet: Dog } = { pet: rex }; const pen: { readonly pet: Animal } = kennel;`,
		"a read-only map":       animals + `const dogs = new Map<string, Dog>(); const all: ReadonlyMap<string, Animal> = dogs;`,
		"the same type":         animals + `const dogs: Dog[] = [rex]; const same: Dog[] = dogs;`,
		"a class instance":      animals + `class Box<Item> { item: Item; constructor(item: Item) { this.item = item; } } const box: Box<Animal> = new Box<Dog>(rex);`,
		"a width-only widening": `const point = { x: 1, y: 2 }; const flat: { x: number } = point;`,
		"a callback argument":   animals + `const dogs: Dog[] = [rex]; dogs.forEach((dog) => dog.bark());`,
		// t1 on #drbrp8c: a union slot is judged whole. Judged member by member, its `null` read as a slot a
		// handler could be written into; every element seen as an Element reported onfullscreenchange that way.
		"a union slot": `interface Handler { readonly run: (label: string) => string } interface SameHandler { readonly run: (label: string) => string } declare const source: { onEvent: Handler | null }; const wide: { onEvent: SameHandler | null } = source;`,
		// t1b on #drbrp8c: what the wider slot can hold is assignable back, so writing it is safe. The two slot
		// types differ (one is readonly), so identity would report it. The readonly is the target's: the source's
		// `readonly name` made writable is shape 1 of #gvzdft9, a hole of its own.
		"a slot assignable back": `declare const source: { pet: { name: string } }; const wide: { pet: { readonly name: string } } = source;`,
		"an any element":         `declare const loose: any[]; const strict: string[] = loose;`,
		"a new array from map":   animals + `const dogs: Dog[] = [rex]; const all: Animal[] = dogs.map((dog) => dog);`,
		"a new map":              animals + `const all: Map<string, Animal> = new Map<string, Dog>([['Rex', rex]]);`,
		"Object.values":          animals + `declare const byName: { readonly [name: string]: Dog }; const all: Animal[] = Object.values(byName);`,
		"a conditional literal":  `declare const flag: boolean; const maybe: { kind: string } | undefined = flag ? { kind: 'Circle' } : undefined;`,
		// #53w68gt's near-misses: a constraint with no writable slot, the same parameter, an unconstrained one,
		// and an intersection whose array member is the same and whose other member is read-only.
		"a type parameter with a readonly constraint":      animals + `function admit<Pack extends readonly Animal[], Narrow extends Pack>(narrow: Narrow): Pack { const pack: Pack = narrow; return pack; }`,
		"a type parameter constrained to a readonly shape": animals + `function admit<Pack extends Animal, Narrow extends Pack>(narrow: Narrow): Pack { return narrow; }`,
		"an unconstrained type parameter":                  `function hold<Item, Other extends Item>(other: Other): Item { return other; }`,
		// #53w68gt's measured false findings, each a pair that is no flow into the parameter.
		"a generic method seen under two receivers": `class Entity { name = ''; clone<T extends this>(): T { return this as T; } } class Session extends Entity { token = ''; } declare const session: Readonly<Session & object>; const wide: Readonly<Session> = session;`,
		"a union member the parameter is not":       `class Tracked { name = ''; } function insert<Entity extends Tracked>(entity: Readonly<Entity> | readonly Entity[]): readonly Entity[] { const entities: readonly Entity[] = Array.isArray(entity) ? entity : [entity]; return entities; }`,
		"a parameter narrowed to non-null":          `function keep<Value extends { items: string[] } | null>(value: Value, set: (next: Value) => void): void { if(value !== null) { set(value); } }`,
		// #b9a0wgy's near-misses: a readonly member, and a brand, whose object member is a phantom.
		"an object intersection with a readonly member": animals + `interface Named { readonly label: string } const kennel: { pet: Dog } & Named = { pet: rex, label: 'kennel' }; const pen: { readonly pet: Animal } & Named = kennel;`,
		"a class constructor's prototype":               `class Form { readonly kind = 'Form'; } type ClassType = (new () => object) & { prototype: object }; const resolve: () => typeof Form = () => Form; const wide: () => ClassType = resolve;`,
		"a branded primitive":                           `declare const brand: unique symbol; type Slot = string & { readonly [brand]?: () => 'value' }; const text: Slot = 'Hello'; const wide: string & { readonly [brand]?: () => string } = text;`,
		"an intersection with the same array":           animals + `declare const tagged: Animal[] & { readonly tag: 'kennel' }; const wide: Animal[] & { readonly tag: string } = tagged;`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, InvariantMutable, source))
		})
	}
}

// TestInvariantMutableFiresWhereAReadOnlySlotIsMadeWritable: shape 1 of #gvzdft9, six-probes.tgz's s1. A
// read-only property is covariant, so it may hold a Dog behind `Animal`, and a writable view of it lets a Cat in;
// tsc 6.0.3 accepts it, and Node throws `bark is not a function`.
func TestInvariantMutableFiresWhereAReadOnlySlotIsMadeWritable(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		span   string
	}{
		"the same type, read-only then writable": {animals + `
const kennel: { readonly pet: Dog } = { pet: rex };
const view: { readonly pet: Animal } = kennel;
const pen: { pet: Animal } = view;`, "view"},
		"one level down": {animals + `
declare const yard: { readonly kennel: { readonly pet: Animal } };
const open: { readonly kennel: { pet: Animal } } = yard;`, "yard"},
		// six-probes.tgz's s1b: a read-only string may hold a literal, so `sounds[dog.kind]` stops being a function.
		"a primitive, under a slot assignable back": {`declare const source: { pet: { readonly name: string } }; const wide: { pet: { name: string } } = source;`, "source"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, InvariantMutable, fixture.source)
			rule_testing.ExpectFindings(t, result, "readonlyMadeWritable")
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestInvariantMutableFiresWhereADestructuringAssignmentWidens: shape 2 of #gvzdft9, six-probes.tgz's s2. Each
// target of an assignment pattern is a slot the value's part is put into, so `[animals] = [dogs]` is
// `animals = dogs`; tsc 6.0.3 accepts it, and Node throws `dogs[1].bark is not a function`.
func TestInvariantMutableFiresWhereADestructuringAssignmentWidens(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		span   string
	}{
		"an array pattern from a literal":  {animals + `const dogs: Dog[] = [rex]; let all: Animal[] = []; [all] = [dogs];`, "dogs"},
		"an array pattern from a tuple":    {animals + `declare const pair: [Dog[], number]; let all: Animal[] = []; [all] = pair;`, "pair"},
		"an object pattern":                {animals + `declare const holder: { pets: Dog[] }; let all: Animal[] = []; ({ pets: all } = holder);`, "holder"},
		"a shorthand in an object pattern": {animals + `declare const holder: { pets: Dog[] }; let pets: Animal[] = []; ({ pets } = holder);`, "holder"},
		"a nested pattern":                 {animals + `const dogs: Dog[] = [rex]; let all: Animal[] = []; [[all]] = [[dogs]];`, "dogs"},
		"an object pattern from a literal": {animals + `const dogs: Dog[] = [rex]; let all: Animal[] = []; ({ pets: all } = { pets: dogs });`, "dogs"},
		// Contextually typed as the literal's own elements, so reported once, by the literal's finder.
		"after a spread in a literal value": {animals + `const dogs: Dog[] = [rex]; declare const cats: Animal[][]; let all: Animal[] = []; [all] = [...cats, dogs];`, "dogs"},
		"a nested pattern from a tuple":     {animals + `declare const pair: [Dog[]]; let all: Animal[] = []; [[all]] = [pair];`, "pair"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, InvariantMutable, fixture.source)
			rule_testing.ExpectFindings(t, result, "mutableWidening")
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestInvariantMutableStaysCleanWhereADestructuringAssignmentCannotWriteThrough: shape 2's near-misses.
func TestInvariantMutableStaysCleanWhereADestructuringAssignmentCannotWriteThrough(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"the same type":     animals + `const dogs: Dog[] = [rex]; let same: Dog[] = []; [same] = [dogs];`,
		"a readonly target": animals + `const dogs: Dog[] = [rex]; let view: readonly Animal[] = []; [view] = [dogs];`,
		"a fresh element":   animals + `let all: Animal[] = []; [all] = [[rex]];`,
		// A rest element is an array the pattern builds, no one else's.
		"a rest element": animals + `const dogs: Dog[] = [rex]; let rest: Animal[] = []; [...rest] = dogs;`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, InvariantMutable, source))
		})
	}
}

// TestInvariantMutableFiresWhereASpreadSharesAMutablePart: shape 3 of #gvzdft9, six-interface.tgz's s3 and s3b. A
// spread copies the top level only, so what the copy's slots hold is the original's, and a wider view writes
// into it; tsc 6.0.3 accepts it, and Node throws `bark is not a function`.
func TestInvariantMutableFiresWhereASpreadSharesAMutablePart(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		span   string
	}{
		"an object spread":                 {animals + `const holder = { pets: [rex] }; const view: { pets: Animal[] } = { ...holder };`, "holder"},
		"an array spread":                  {animals + `const pens: Dog[][] = [[rex]]; const all: Animal[][] = [...pens];`, "pens"},
		"beside a property written after":  {animals + `const holder = { pets: [rex], name: 'kennel' }; const view: { pets: Animal[]; name: string } = { ...holder, name: 'yard' };`, "holder"},
		"a readonly property still shares": {animals + `const holder = { pets: [rex] }; const view: { readonly pets: Animal[] } = { ...holder };`, "holder"},
		// Fresh elements still hold what they were given: dogs is shared.
		"a fresh mapped element holding a shared array":       {animals + `const dogs: Dog[] = [rex]; const names = ['a']; const all: { pets: Animal[] }[] = [...names.map((name) => ({ pets: dogs }))];`, "names.map((name) => ({ pets: dogs }))"},
		"a map returning what someone holds":                  {animals + `const kennels: { pets: Dog[] }[] = [{ pets: [rex] }]; const all: { pets: Animal[] }[] = [...kennels.map((kennel) => kennel)];`, "kennels.map((kennel) => kennel)"},
		"a map returning what someone holds, at its own slot": {`const kennels: { kind: 'dog' }[] = [{ kind: 'dog' }]; const all: { kind: string }[] = [...kennels.map((kennel) => kennel)];`, "kennels.map((kennel) => kennel)"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, InvariantMutable, fixture.source)
			rule_testing.ExpectFindings(t, result, "mutableWidening")
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestInvariantMutableStaysCleanWhereASpreadSharesNothingWritable: shape 3's near-misses.
func TestInvariantMutableStaysCleanWhereASpreadSharesNothingWritable(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		// s3c: written again after the spread, so the copy holds the later value.
		"a property written after":   animals + `const holder = { pets: [rex] }; const view: { pets: Animal[] } = { ...holder, pets: [] };`,
		"a later spread writing it":  animals + `const holder = { pets: [rex] }; const fresh: { pets: Animal[] } = { pets: [] }; const view: { pets: Animal[] } = { ...holder, ...fresh };`,
		"a read-only array":          animals + `const holder = { pets: [rex] }; const view: { pets: readonly Animal[] } = { ...holder };`,
		"the same type":              animals + `const holder: { pets: Animal[] } = { pets: [rex] }; const view: { pets: Animal[] } = { ...holder };`,
		"a fresh spread":             animals + `const view: { pets: Animal[] } = { ...{ pets: [rex] } };`,
		"an array spread of objects": animals + `const dogs: Dog[] = [rex]; const all: Animal[] = [...dogs];`,
		// A map whose callback returns fresh literals builds its elements too: nobody else holds them.
		"a spread of fresh mapped elements": `interface Row { section: 'doctrine' | 'soul' } const pages = ['a']; const rows: Row[] = [...pages.map((page) => ({ section: 'doctrine' as const }))];`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, InvariantMutable, source))
		})
	}
}

// TestInvariantMutableFiresWhereAMethodTakesANarrowerParameter: shape 5 of #gvzdft9, six-interface.tgz's s5 and
// s5b. tsc compares a method's parameters in both directions, so a method taking only a Dog passes as one taking
// any Animal; tsc 6.0.3 accepts it, and a Cat passed through the wider type is `bark is not a function` in Node.
func TestInvariantMutableFiresWhereAMethodTakesANarrowerParameter(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		span   string
	}{
		"a literal's method":           {animals + `interface Sink { put(animal: Animal): void } const sink: Sink = { put(dog: Dog) { dog.bark(); } };`, "put"},
		"a held object's method":       {animals + `interface Sink { put(animal: Animal): void } declare const dogSink: { put(dog: Dog): void }; const sink: Sink = dogSink;`, "dogSink"},
		"a function where a method is": {animals + `interface Sink { put(animal: Animal): void } declare const dogSink: { put: (dog: Dog) => void }; const sink: Sink = dogSink;`, "dogSink"},
		// #bbtfx99, generic-probe.tgz: a generic source is instantiated in the target's context, as tsc relates it, and
		// what it takes is then a Dog. g3, g6 and g5; each is accepted by tsc 6.0.3 and a TypeError in Node.
		"a generic method":                       {animals + `interface Sink { put(animal: Animal): void } declare const dogSink: { put<T extends Dog>(dog: T): void }; const sink: Sink = dogSink;`, "dogSink"},
		"a class's generic method":               {animals + `interface Sink { put(animal: Animal): void } class Kennel { put<T extends Dog>(dog: T): void { dog.bark(); } } const sink: Sink = new Kennel();`, "new Kennel()"},
		"two generic methods, a plain parameter": {animals + `interface Sink { put<T>(animal: Animal, tag: T): void } declare const dogSink: { put<U>(dog: Dog, tag: U): void }; const sink: Sink = dogSink;`, "dogSink"},
		// A default takes `undefined` as well, and still not a Cat.
		"a defaulted parameter": {animals + `interface Sink { put(animal?: Animal): void } class Kennel { put(dog: Dog = rex): void { dog.bark(); } } const sink: Sink = new Kennel();`, "new Kennel()"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, InvariantMutable, fixture.source)
			rule_testing.ExpectFindings(t, result, "methodParameterNarrowed")
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestInvariantMutableFiresWhereAMethodReturnsWhatItHolds: #y0ejf6a, return-probe.tgz's r1 to r4. A method hands back
// a container its object still holds, so a Cat pushed through the wider return fills it; tsc 6.0.3 accepts each, and
// Node throws `bark is not a function`.
func TestInvariantMutableFiresWhereAMethodReturnsWhatItHolds(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		span   string
	}{
		"a class's field":       {animals + `interface Pound { all(): Animal[] } class Kennel { dogs: Dog[] = [rex]; all(): Dog[] { return this.dogs; } } const pound: Pound = new Kennel();`, "new Kennel()"},
		"a literal's method":    {animals + `interface Pound { all(): Animal[] } const dogs: Dog[] = [rex]; const pound: Pound = { all() { return dogs; } };`, "all"},
		"a map":                 {animals + `interface Registry { byName(): Map<string, Animal> } class Kennel { names = new Map<string, Dog>(); byName(): Map<string, Dog> { return this.names; } } const registry: Registry = new Kennel();`, "new Kennel()"},
		"an interface's method": {animals + `interface Pound { all(): Animal[] } declare const dogPound: { all(): Dog[] }; const pound: Pound = dogPound;`, "dogPound"},
		// One return that builds and one that does not: the method can hand back what it holds.
		"one return of two held": {animals + `interface Pound { all(fresh: boolean): Animal[] } class Kennel { dogs: Dog[] = [rex]; all(fresh: boolean): Dog[] { if(fresh) { return []; } return this.dogs; } } const pound: Pound = new Kennel();`, "new Kennel()"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, InvariantMutable, fixture.source)
			rule_testing.ExpectFindings(t, result, "mutableWidening")
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestInvariantMutableStaysCleanWhereAMethodTakesWhatItIsGiven: shape 5's near-misses.
func TestInvariantMutableStaysCleanWhereAMethodTakesWhatItIsGiven(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		// s5c: a wider parameter takes whatever the view passes.
		"a wider parameter":   animals + `interface DogSink { put(dog: Dog): void } const sink: DogSink = { put(animal: Animal) { animal.name; } };`,
		"the same parameter":  animals + `interface Sink { put(animal: Animal): void } const sink: Sink = { put(animal: Animal) { animal.name; } };`,
		"no parameter":        `interface Clock { now(): number } const clock: Clock = { now() { return 1; } };`,
		"a held wider method": animals + `interface DogSink { put(dog: Dog): void } declare const anySink: { put(animal: Animal): void }; const sink: DogSink = anySink;`,
		// #bbtfx99: two generic signatures are related by instantiating one in the other's context, so one `T` is never
		// paired with another. Promise's `then` is the same `then` under two promises; g4 and n2 were false findings on
		// main, a literal's generic method paired with its target's as they stood, and both run clean in Node.
		"Promise's then":               animals + `const dogs: Promise<Dog> = Promise.resolve(rex); const all: Promise<Animal> = dogs;`,
		"a generic identity method":    animals + `interface Sink { put<T extends Animal>(animal: T): T } const sink: Sink = { put<T extends Animal>(animal: T): T { return animal; } };`,
		"a wider generic method":       animals + `interface DogSink { put<T extends Dog>(dog: T): void } const sink: DogSink = { put<T extends Animal>(animal: T) { animal.name; } };`,
		"an unconstrained generic one": animals + `interface Sink { put(animal: Animal): void } declare const echo: { put<T>(value: T): void }; const sink: Sink = echo;`,
		// Inferred as a Dog from the target, as tsc infers it, so the callback takes a Dog on both sides. Read at its
		// constraint instead, it would take an Animal, and the target's callback taking a Dog would look narrower.
		"an inferred type argument": animals + `interface DogSink { put(dog: Dog, then: (dog: Dog) => void): void } declare const sink: { put<T extends Animal>(animal: T, then: (animal: T) => void): void }; const dogSink: DogSink = sink;`,
		// A generic function's two signatures are related the same way, so its Promise<Row[]> is one `Row` on both sides
		// when it reaches `then` (DrizzleAdapterSql's executeRows on api).
		"a generic function property": `interface Adapter { executeRows: <Row>(query: string) => Promise<Row[]> } class Sql { executeRows<Row>(query: string): Promise<Row[]> { return Promise.resolve([]); } } const adapter: Adapter = new Sql();`,
		// The parameter is judged whole: below `then`'s callback is the promise's value, here a literal nobody else holds.
		"a promise's value": `const done: Promise<{ result: string }> = Promise.resolve({ result: 'ok' as const });`,
		// A default takes what the wider type's `?` passes (OrmDatabase's increment on api).
		"a defaulted parameter": `interface Counter { add(value?: number): void } class Tally { add(value = 1): void { value; } } const counter: Counter = new Tally();`,
		// #y0ejf6a's near-misses, return-probe.tgz's n1 to n4: a read-only return, a method building what it returns, and
		// the library's iterator, whose next() hands back a new result each call (276 findings on the consumers when a
		// method's return was first paired).
		"a readonly return":              animals + `interface Pound { all(): readonly Animal[] } class Kennel { dogs: Dog[] = [rex]; all(): Dog[] { return this.dogs; } } const pound: Pound = new Kennel();`,
		"a copy returned":                animals + `interface Pound { all(): Animal[] } class Kennel { dogs: Dog[] = [rex]; all(): Dog[] { return this.dogs.slice(); } } const pound: Pound = new Kennel();`,
		"a literal returned":             animals + `interface Pound { all(): Animal[] } class Kennel { all(): Dog[] { return [rex]; } } const pound: Pound = new Kennel();`,
		"a literal returned, or nothing": animals + `interface Pound { all(): Animal[] | undefined } class Kennel { all(): Dog[] | undefined { return [rex]; } } const pound: Pound = new Kennel();`,
		// #jpdf48x, function-probe.tgz's n1 to n3: a function building what it returns, every call, as an arrow, a copy
		// and a declaration's body. Each runs clean in Node and was reported on main.
		"an arrow's literal": animals + `interface Pound { all(): Animal[] } const pound: Pound = { all: () => [rex] };`,
		"a function's copy":  animals + `const dogs: Dog[] = [rex]; const dogsOf = (): Dog[] => dogs.slice(); const animalsOf: () => Animal[] = dogsOf;`,
		// Built all the way down: every level of the literal is written in place (nexus's validators on the consumers).
		"a nested literal": `interface Result { valid: string; errors: { identifier: string }[] } const validate = (): { valid: 'no'; errors: { identifier: 'short' }[] } => ({ valid: 'no', errors: [{ identifier: 'short' }] }); const wide: () => Result = validate;`,
		// Written after the spread, so the literal holds the later value there (MessageService's reconnection data on api).
		"a literal written after a spread": `declare const base: { other: number }; const data = (): { other: number; mode: { id: 'x' } } => ({ ...base, mode: { id: 'x' } }); const wide: () => { other: number; mode: { id: string } } = data;`,
		// #qq4haam, const-probe.tgz's n1: a local value built here and only returned (TableColumnFilterGroup on the
		// consumers, its `filters` a `.map` result), through a shorthand and a plain property.
		"a local new array returned": animals + `const kennel: Dog[] = [rex]; const make = (): { pets: Dog[] } => { const pets = kennel.map((dog) => dog); return { pets }; }; const wide: () => { pets: Animal[] } = make;`,
		"a local literal returned":   animals + `const make = (): { pets: Dog[] } => { const pets = [rex]; return { pets: pets }; }; const wide: () => { pets: Animal[] } = make;`,
		// Another binding of the same name is no reference to this one, which the checker tells apart.
		"a shadowing name elsewhere": animals + `const make = (): { pets: Dog[] } => { const pets = [rex]; const count = (pets: Dog[]): number => pets.length; count([]); return { pets }; }; const wide: () => { pets: Animal[] } = make;`,
		// TableColumnFilterGroup's shape whole: the local is a `.map` whose callback builds each element, so an element's
		// own slots are no one else's either.
		"a local map building its elements": `interface Filter { id?: string; name: string } declare const incoming: { filters?: Filter[] }; const make = (): { filters?: { id: string; name: string }[] } => { const filters = incoming.filters?.map(function(filter) { return { ...filter, id: filter.id ?? 'x' }; }); return { filters }; }; const wide: () => { filters?: Filter[] } = make;`,
		"a spread with a local after it":    animals + `declare const incoming: { name: string; pets?: Dog[] }; const make = (): { name: string; pets?: Dog[] } => { const pets = incoming.pets?.map((dog) => dog); return { ...incoming, pets }; }; const wide: () => { name: string; pets?: Animal[] } = make;`,
		"a declaration's literals":          animals + `function makeDogs(): Dog[] { if(rex.name === '') { return []; } return [rex]; } const make: () => Animal[] = makeDogs;`,
		"a literal's method building":       animals + `interface Pound { all(): Animal[] } const pound: Pound = { all() { return [rex]; } };`,
		"an iterator's next()":              animals + `declare const dogs: Iterator<Dog>; const all: Iterator<Animal> = dogs;`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, InvariantMutable, source))
		})
	}
}

// classAnimals is the vocabulary with classes, which nominal-class judges as values and this rule as what a slot
// holds (#5y54eyj).
const classAnimals = `
class Animal { name: string; constructor(name: string) { this.name = name; } }
class Dog extends Animal { bark(): string { return this.name; } }
class Box<Item> { item: Item; constructor(item: Item) { this.item = item; } }
declare const rex: Dog;
`

// TestInvariantMutableFiresWhereASlotHoldsAClassInstance: #5y54eyj, class-probe.tgz. nominal-class accepts a Dog as an
// Animal, rightly for the value, so the slot holding it is this rule's; tsc 6.0.3 accepts each, and Node throws
// `bark is not a function`.
func TestInvariantMutableFiresWhereASlotHoldsAClassInstance(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		span   string
		id     string
	}{
		// plain-class-array
		"an array's element": {classAnimals + `const dogs: Dog[] = [rex]; const all: Animal[] = dogs;`, "dogs", "mutableWidening"},
		// c2-property
		"a mutable property": {classAnimals + `const kennel: { pet: Dog } = { pet: rex }; const pen: { pet: Animal } = kennel;`, "kennel", "mutableWidening"},
		// c3-map
		"a map's value": {classAnimals + `const dogs = new Map<string, Dog>(); const all: Map<string, Animal> = dogs;`, "dogs", "mutableWidening"},
		// c5-readonly-made-writable
		"a read-only slot made writable": {classAnimals + `declare const view: { readonly pet: Animal }; const pen: { pet: Animal } = view;`, "view", "readonlyMadeWritable"},
		// c4-method-parameter
		"a method's parameter": {classAnimals + `interface Sink { put(animal: Animal): void } const sink: Sink = { put(dog: Dog) { dog.bark(); } };`, "put", "methodParameterNarrowed"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := runAdamic(t, InvariantMutable, fixture.source)
			rule_testing.ExpectFindings(t, result, fixture.id)
			expectSpans(t, fixture.source, result, fixture.span)
		})
	}
}

// TestInvariantMutableStaysCleanWhereAClassInstanceIsOnlyRead: #5y54eyj's near-misses, and the pair nominal-class
// reports itself.
func TestInvariantMutableStaysCleanWhereAClassInstanceIsOnlyRead(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		// n1, n2, n3
		"a readonly array":    classAnimals + `const dogs: Dog[] = [rex]; const all: readonly Animal[] = dogs;`,
		"the same class":      classAnimals + `const dogs: Dog[] = [rex]; const same: Dog[] = dogs;`,
		"a fresh array":       classAnimals + `const all: Animal[] = [rex];`,
		"a readonly property": classAnimals + `const kennel: { pet: Dog } = { pet: rex }; const pen: { readonly pet: Animal } = kennel;`,
		// n4: Box<Dog> is no Box<Animal>, which is nominal-class's finding on the same pair.
		"another type argument": classAnimals + `const boxes: Box<Dog>[] = [new Box(rex)]; const all: Box<Animal>[] = boxes;`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, InvariantMutable, source))
		})
	}
}

// TestInvariantMutableStaysCleanWhereAReadOnlySlotStaysReadOnly: the near-misses of shape 1.
func TestInvariantMutableStaysCleanWhereAReadOnlySlotStaysReadOnly(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"read-only both sides":        animals + `declare const kennel: { readonly pet: Animal }; const view: { readonly pet: Animal } = kennel;`,
		"a literal type":              `declare const circle: { readonly kind: 'Circle' }; const shape: { kind: 'Circle' } = circle;`,
		"null":                        `declare const empty: { readonly next: null }; const node: { next: null } = empty;`,
		"a writable source":           animals + `declare const kennel: { pet: Animal }; const pen: { pet: Animal } = kennel;`,
		"a fresh literal of the view": animals + `const pen: { pet: Animal } = { pet: rex };`,
		// s1c: a library getter throws on the write before anything is stored.
		"a library getter": `const pattern = /dog/; const view: { source: string } = pattern;`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAdamic(t, InvariantMutable, source))
		})
	}
}

// TestInvariantMutableStaysCleanOnAdamicsPrograms: Adamic 0.1 compiles all ten.
func TestInvariantMutableStaysCleanOnAdamicsPrograms(t *testing.T) {
	t.Parallel()
	expectProgramsClean(t, InvariantMutable)
}

// TestInvariantMutableNamesTheSlot: the message says which part can be written through and both types
// there, since `'Dog[]' is seen as 'Animal[]'` alone does not say where to put the `readonly`.
func TestInvariantMutableNamesTheSlot(t *testing.T) {
	t.Parallel()
	source := animals + `
const kennel: { readonly dogs: Dog[] } = { dogs: [rex] };
const pen: { readonly dogs: Animal[] } = kennel;`
	result := runAdamic(t, InvariantMutable, source)
	rule_testing.ExpectFindings(t, result, "mutableWidening")
	description := result.Diagnostics[0].Message.Description
	for _, want := range []string{"`kennel.dogs[]`", "'Dog'", "'Animal'"} {
		if !strings.Contains(description, want) {
			t.Errorf("the message does not name %s: %s", want, description)
		}
	}
}
