// a-check: refused a value of type T[] seen as {}[], which can write {} where T is read
// TypeScript 6.0.3 core.ts helpers; standalone oracle fixture.
export function append<T extends {}>(to: T[], value: T | undefined): T[];
/** @internal */
export function append<T extends {}>(to: T[] | undefined, value: T): T[];
/** @internal */
export function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined;
export function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined {
    if (value === undefined) return to;
    if (to === undefined) return [value];
    to.push(value);
    return to;
}

export function isArray(value: unknown): value is readonly unknown[] {
    // See: https://github.com/microsoft/TypeScript/issues/17002
    return Array.isArray(value);
}

function concatenate<T>(array1: T[], array2: T[]): T[] { return [...array1, ...array2]; }

function probe<T extends {}>(xs: T | T[] | undefined, ys: T | T[] | undefined) {
 if (xs === undefined) return ys;
 if (ys === undefined) return xs;
 if (isArray(xs) && !isArray(ys)) { const result = append(xs, ys); console.log(typeof result[0]); return result; }
 return undefined;
}

const result = probe<{ n: number }>([{ n: 11 }], { n: 31 });
if (isArray(result)) {
 const first = result[0]; if (first !== undefined) console.log(String(first.n + 7));
 for (const item of result) { console.log(String(item.n)); console.log(typeof item); }
}
