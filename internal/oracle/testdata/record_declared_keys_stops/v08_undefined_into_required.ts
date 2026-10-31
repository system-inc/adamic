interface Counts { required: number; [k: string]: number | undefined; }
const keys: string[] = ["required", "other"];
const key = keys[0] ?? "other";
const counts: Counts = { required: 1 };
counts[key] = undefined;
console.log(typeof counts.required);
