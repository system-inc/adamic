package lower

import (
	"strings"
	"testing"
)

// genericMethodAnimals are the classes every generic-method view below is written against: a Dog
// method handed a Cat reads bark off something that has none, which Node throws on.
const genericMethodAnimals = `class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
`

// A generic method seen through a view whose method isn't its own type stops, unjudged, on either side
// and through an interface (#q6gs98r, Language's probes g1, g4, g5, g6 and g7 on #7fd51gk, and o2, an
// overloaded view, from her attack on #q6gs98r). tsc relates a
// method's parameters both ways, so each of these is accepted there and hands a Dog method a Cat.
func TestAGenericMethodSeenThroughAnotherViewIsNotYet(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source string }{
		{"g1, T unused on both sides", `interface Handler { handle<T>(x: Animal): string; }
class DogHandler { handle<T>(x: Dog): string { return x.bark(); } }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat()));
`},
		{"g4, T only in a callback's type", `interface Handler { handle<T>(x: Animal, f: (t: T) => string): string; }
class DogHandler { handle<T>(x: Dog, f: (t: T) => string): string { return x.bark(); } }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat(), (t: number) => String(t)));
`},
		{"g5, generic on the source side only", `interface Handler { handle(x: Animal): string; }
class DogHandler { handle<T>(x: Dog): string { return x.bark(); } }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat()));
`},
		{"g6, generic on the target side only", `interface Handler { handle<T>(x: Animal): string; }
class DogHandler { handle(x: Dog): string { return x.bark(); } }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat()));
`},
		{"g7, an interface seen as another", `interface Handler { handle<T>(x: Animal): string; }
interface DogHandler { handle<T>(x: Dog): string; }
class Real { handle<T>(x: Dog): string { return x.bark(); } }
const d: DogHandler = new Real();
const h: Handler = d;
console.log(h.handle(new Cat()));
`},
		{"o2, an overloaded generic view", `class DogHandler { handle<T>(x: Dog): string { return x.bark(); } }
interface Handler {
	handle<T>(x: Animal): string;
	handle<T>(x: Animal, y: number): string;
}
const h: Handler = new DogHandler();
console.log(h.handle(new Cat()));
`},
		{"a host method a view makes generic", `const p: { chdir<T>(directory: string | number): void } = process;
p.chdir(5);
`},
		{"a host method a view makes generic, exit", `const p: { exit<T>(code?: number | string | null | boolean): void } = process;
p.exit(true);
`},
		{"one class's generic method under two type arguments", `class Box<E> { put<T>(x: E): string { return "kept"; } }
const dogs: Box<Dog> = new Box<Dog>();
const animals: Box<Animal> = dogs;
console.log(animals.put(new Cat()));
`},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, genericMethodAnimals+row.source, extension)
				if err == nil {
					t.Fatal("lowered, but a Dog method is handed a Cat through this view")
				}
				if !strings.Contains(err.Error(), "stage 0 can't lower a generic method ") || !strings.Contains(err.Error(), " seen as ") {
					t.Fatalf("got %q, want the generic method named as NotYet", err)
				}
			})
		}
	}
}

// A generic method whose type is identical to the view's is the same method, so it's still judged as
// it was: Language's control g2, and a class seen through the interface it implements.
func TestAGenericMethodSeenAsItsIdenticalTypeLowers(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source string }{
		{"g2, an honest handle<T>(x: Animal)", `interface Handler { handle<T>(x: Animal): string; }
class AnyHandler { handle<T>(x: Animal): string { return x.name; } }
const h: Handler = new AnyHandler();
console.log(h.handle(new Cat()));
`},
		{"a class seen through the interface it implements", `interface Handler { handle<T>(x: Animal): string; }
class AnyHandler implements Handler { handle<T>(x: Animal): string { return x.name; } }
const h: Handler = new AnyHandler();
console.log(h.handle(new Dog()));
`},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				if _, err := lowerSourceAs(t, genericMethodAnimals+row.source, extension); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
