type Point = { a: number; b: number };
type Bag = Record<string, number>;
function fill(bag: Bag): string { bag["c"] = 3; return String(bag["a"]) + " " + String(Object.keys(bag).length); }
const fixed: Point = { a: 1, b: 2 };
function* bags(): Generator<Bag, void, unknown> { yield fixed; }
for (const bag of bags()) console.log(fill(bag));
