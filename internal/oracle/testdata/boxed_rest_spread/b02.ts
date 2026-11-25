interface SortedReadonlyArray<T> extends ReadonlyArray<T> { " __sortedArrayBrand": undefined; }
interface SortedArray<T> extends Array<T> { " __sortedArrayBrand": undefined; }
const values: (number | string)[] = [0, "one"];
function target(...rest: SortedArray<number | string>): void {
    console.log(String(rest.length));
    for (const value of rest) console.log(String(value));
}
target(...values);
