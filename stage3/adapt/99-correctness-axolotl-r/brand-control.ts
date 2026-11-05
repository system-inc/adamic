interface SortedReadonlyArray<T> extends ReadonlyArray<T> { " __sortedArrayBrand": undefined; }
function mint(xs: number[]): SortedReadonlyArray<number> {
    return xs as unknown as SortedReadonlyArray<number>;
}
console.log(String(mint([1, 2]).length));
