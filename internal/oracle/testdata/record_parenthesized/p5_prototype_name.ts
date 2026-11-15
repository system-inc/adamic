// A parenthesized record read naming an Object.prototype member stays refused at the access inside (#grc95d6).
const r: Record<string, number> = { k: 2 };
console.log(String((r["toString"])));
