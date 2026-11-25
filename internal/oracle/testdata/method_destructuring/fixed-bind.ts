interface G { add(b: number): number }
function add(a: number, b: number): number { return a + b; }
const g: G = { add: add.bind(undefined, 1) };
const { add: increment } = g;
console.log(String(increment(2)));
