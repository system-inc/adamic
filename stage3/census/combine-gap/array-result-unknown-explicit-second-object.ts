// a-check: checked
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

function second<T extends {}>(xs: T[], ys: T): T[] { return append<T>(xs, ys); }
function probe<T extends {}>(xs: T | T[] | undefined, ys: T | T[] | undefined): T[] | undefined {
 if (xs === undefined) return undefined;
 if (ys === undefined) return undefined;
 if (isArray(xs) && !isArray(ys)) { return second<T>(xs, ys); }
 return undefined;
}

const result = probe<{ n: number }>([{ n: 11 }], { n: 31 });
if (isArray(result)) {
 for (const item of result) { console.log(String(item.n)); console.log(typeof item); }
}
