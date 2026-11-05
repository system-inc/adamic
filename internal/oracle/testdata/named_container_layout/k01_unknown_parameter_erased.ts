function show(items: readonly unknown[]): void {
    for (const x of items) { console.log(String(x)); }
}
const nums: number[] = [0, 1];
show(nums);
