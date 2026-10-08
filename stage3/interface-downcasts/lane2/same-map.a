// TypeScript 6.0.3 core.ts sameMap. Index assertions preserve the loop-bound proof.
// Both generic array assertion chains are unchanged, including the nullable return.
// The row-2 slice cast is unchanged; overload admission is tested separately.
export function sameMap<T, U = T>(array: readonly T[] | undefined, f: (x: T, i: number) => U): readonly U[] | undefined {
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const item = array[i]!;
            const mapped = f(item, i);
            if (item as unknown !== mapped) {
                const result: U[] = array.slice(0, i) as unknown[] as U[];
                result.push(mapped);
                for (i++; i < array.length; i++) {
                    result.push(f(array[i]!, i));
                }
                return result;
            }
        }
    }
    return array as unknown[] as U[];
}
const original = [1, 2, 3];
const unchanged = sameMap(original, (x: number) => x);
const changed = sameMap(original, (x: number, i: number) => i === 1 ? 20 : x)!;
console.log(`${unchanged === original}:${changed[0]}:${changed[1]}:${changed[2]}:${sameMap<number>(undefined, (x: number) => x) === undefined}`);
