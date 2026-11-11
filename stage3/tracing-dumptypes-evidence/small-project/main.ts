export type Choice = { kind: "left"; value: string } | { kind: "right"; value: number };
export type Selected<T> = T extends { value: infer V } ? V : never;
export const pair: [Choice, boolean] = [{ kind: "left", value: "trace" }, true];
export function map<T, U>(xs: readonly T[], f: (x: T) => U): U[] { return xs.map(f); }
export const result = map([1, 2], x => ({ value: x, label: String(x) }));
