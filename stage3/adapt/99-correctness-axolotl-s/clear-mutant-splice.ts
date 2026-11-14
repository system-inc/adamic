export function clear<T>(array: T[]): void {
    array.splice(0);
}
const values: number[] = [1, 2, 3];
clear(values);
