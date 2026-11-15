// A parenthesized record read as the receiver of ?. (#grc95d6).
const r: Record<string, number[]> = { k: [1] };
console.log(String((r["k"])?.length));
