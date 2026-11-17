




































































































































































































































































/** @internal */

/** @internal */

/** @internal */

/** @internal */

/** @internal */

/** @internal */
export function filter<T, U extends T>(array: readonly T[] | undefined, f: (x: T) => x is U): readonly U[] | undefined;
/** @internal */
export function filter<T>(array: readonly T[] | undefined, f: (x: T) => boolean): readonly T[] | undefined;
/** @internal */
export function filter<T>(array: readonly T[] | undefined, f: (x: T) => boolean): readonly T[] | undefined {
    if (array !== undefined) {
        const len = array.length;
        let i = 0;
        while (i < len && f(array[i]!)) i++;
        if (i < len) {
            const result = array.slice(0, i);
            i++;
            while (i < len) {
                const item = array[i]!;
                if (f(item)) {
                    result.push(item);
                }
                i++;
            }
            return result;
        }
    }
    return array;
}
