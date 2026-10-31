interface Counts { required: number; [k: string]: number | undefined; }
const keys: string[] = ["required", "other"];
const key = keys[0] ?? "other";
const counts: Counts = { required: 1 };
const maybe: number | undefined = keys.length > 5 ? 1 : undefined;
counts[key] = maybe;
console.log(typeof counts.required);
