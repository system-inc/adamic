// tsc's four sorted-array brand constructions (core.ts:736, 759, 764, 1039), spelled `as unknown as`
// as ruled on #b5w3ycg, and a fresh literal minted into a branded array of a union element (Lower's
// pin beside #b8fwyfx's container fit). Each mint is erased: the value is the array it was made from,
// and the brand member is never present on Node or natively.
interface SortedReadonlyArray<T> extends ReadonlyArray<T> { " __sortedArrayBrand": undefined; }
interface SortedArray<T> extends Array<T> { " __sortedArrayBrand": undefined; }
const emptyArray: never[] = [];
function deduplicateSorted<T>(array: SortedReadonlyArray<T>): SortedReadonlyArray<T> {
    if (array.length === 0) return emptyArray as unknown as SortedReadonlyArray<T>;
    const deduplicated: T[] = [array[0]];
    for (let index = 1; index < array.length; index++) {
        if (array[index] !== deduplicated[deduplicated.length - 1]) deduplicated.push(array[index]);
    }
    return deduplicated as unknown as SortedReadonlyArray<T>;
}
function createSortedArray<T>(): SortedArray<T> { return [] as unknown as SortedArray<T>; }
function toSorted<T>(array: readonly T[], compare: (a: T, b: T) => number): SortedReadonlyArray<T> { return array.slice().sort(compare) as unknown as SortedReadonlyArray<T>; }
const sorted = toSorted([3, 1, 2, 1], (a, b) => a - b);
const fresh = createSortedArray<number>();
fresh.push(4);
const mixed = [0, 1] as unknown as SortedArray<number | string>;
mixed.push('two');
console.log(deduplicateSorted(sorted).join(','));
console.log(String(deduplicateSorted(createSortedArray<number>()).length));
console.log(fresh.join(','));
console.log(mixed.join(','));
console.log(String(sorted.length) + ' ' + String(Array.isArray(sorted)));
