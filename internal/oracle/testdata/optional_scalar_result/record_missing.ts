// A key the record lacks: its type holds no undefined, so only ?.'s test sees it's missing, and the
// call reads undefined, never false or a number.
const counts: Record<string, number[]> = { k: [1] };
console.log(String(counts["zz"]?.push(2)));
console.log(String(counts["zz"]?.includes(1)));
console.log(String(counts["zz"]?.indexOf(1)));
console.log(String(counts["k"]?.includes(1)));
console.log(String(counts["k"]?.indexOf(1)));
console.log(String(counts["k"]?.push(2)));
