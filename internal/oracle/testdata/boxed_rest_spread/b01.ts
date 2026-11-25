interface SortedReadonlyArray<T> extends ReadonlyArray<T> { " __sortedArrayBrand": undefined; }
interface SortedArray<T> extends Array<T> { " __sortedArrayBrand": undefined; }
const values = [0, "one"] as unknown as SortedArray<number | string>;
function target(...rest: (number | string)[]): void {
    console.log(String(rest.length));
    for (const value of rest) console.log(String(value));
}
target(...values);
