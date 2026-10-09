function narrowed(value: number | string | undefined): number {
    if (typeof value === 'number') return value!;
    return 1;
}
console.log(String(narrowed(0)));
