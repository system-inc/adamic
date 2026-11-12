type Point = { a: number; b: number };
type Bag = Record<string, number>;
function fill(bag: Bag): string { bag["c"] = 3; return String(bag["a"]) + " " + String(Object.keys(bag).length); }
const fixed: Point = { a: 1, b: 2 };
function grow<T extends Bag>(bag: T): string { return fill(bag); }
const g = grow<Point>;
console.log(g(fixed));
