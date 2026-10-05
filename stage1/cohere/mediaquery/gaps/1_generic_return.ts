// Gap 1: a generic function isn't instantiated per call, so one whose return type is its type
// parameter is refused. docs/0.1.md compiles generics by monomorphization; generic classes are.
function identity<Item>(item: Item): Item {
	return item;
}
console.log(identity('a'));
