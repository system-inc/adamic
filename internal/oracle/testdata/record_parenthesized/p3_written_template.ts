// A parenthesized record read under ?. in a template, its key written first (#grc95d6).
const r: Record<string, number[]> = {};
r["k"] = [1];
console.log(`${(r["k"])?.length}`);
