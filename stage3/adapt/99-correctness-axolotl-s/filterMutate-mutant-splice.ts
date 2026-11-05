export function filterMutate<T>(array: T[], f: (x: T, i: number, array: T[]) => boolean): void {
    let outIndex = 0;
    for (let i = 0; i < array.length; i++) {
        if (f(array[i]!, i, array)) {
            array[outIndex] = array[i]!;
            outIndex++;
        }
    }
    array.splice(outIndex);
}
const values: number[] = [1, 2, 3];
filterMutate(values, value => value > 1);
