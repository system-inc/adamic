/** @internal */
export function filterNarrowed<T, U extends T>(array: readonly T[], f: (x: T) => x is U): U[] {
    const result: U[] = [];
    for (const value of array) {
        if (f(value)) {
            result.push(value);
        }
    }
    return result;
}
/** @internal */
export function filterNarrowedOptional<T, U extends T>(array: readonly T[] | undefined, f: (x: T) => x is U): U[] | undefined {
    if (array === undefined) return undefined;
    const result: U[] = [];
    for (const value of array) {
        if (f(value)) {
            result.push(value);
        }
    }
    return result;
}
function isString(value: string | number): value is string { return typeof value === "string"; }
const input: (string | number)[] = ["one", 2];
export const first = filterNarrowed(input, isString);
export const second = filterNarrowedOptional(input, isString);
export const third = filterNarrowedOptional<string | number, string>(undefined, isString);
