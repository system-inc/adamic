type Point = { a: number; b: number };
type Bag = Record<string, number>;
function fill(bag: Bag): string { bag["c"] = 3; return String(bag["a"]) + " " + String(Object.keys(bag).length); }
const fixed: Point = { a: 1, b: 2 };
const bag = fixed;
const holder: { bag: Bag } = { bag };
console.log(fill(holder.bag));
