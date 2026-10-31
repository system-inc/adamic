enum Kind { A, B }
interface Tagged { kind: Kind; [k: string]: number; }
const keys: string[] = ["kind", "other"];
const key = keys[0] ?? "other";
const tagged: Tagged = { kind: Kind.A };
const seven: number = keys.length + 5;
tagged[key] = seven;
console.log(String(tagged.kind));
