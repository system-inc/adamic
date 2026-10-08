// A body-proven predicate does not by itself certify array-wide callback effects.
function every(array: readonly (number | string)[] | undefined, callback: (element: number | string, index: number) => element is number): array is readonly number[] | undefined {
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            if (!callback(array[i]!, i)) return false;
        }
    }
    return true;
}
const values: (number | string)[] = [1, 2];
function isNumber(value: number | string, index: number): value is number {
    values[0] = "changed";
    return typeof value === "number";
}
console.log(`${every(values, isNumber)}:${typeof values[0]}`);
