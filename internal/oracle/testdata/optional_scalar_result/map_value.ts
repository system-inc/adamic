// A Map's value asserted present, then pushed through ?.
const m: Map<string, number[]> = new Map([["k", [1]]]);
const got = m.get("k")!;
console.log(String(got?.push(2)));
