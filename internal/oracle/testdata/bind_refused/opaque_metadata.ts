// Node prints 1. The existing graph/dynamic boundary must diagnose the read.
function join(a: string, b: string): string { return a + b; }
const bound = join.bind(undefined, "fixed");
function observe(value: Function): number { return value.length; }
console.log(String(observe(bound)));
