function show(value: { a: number }): string {
    return JSON.stringify(value);
}
const extra = { b: 2 };
console.log(show({ ...extra, a: 1 }));
