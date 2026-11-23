function show(value: { a: number }): string {
    return JSON.stringify(value) ?? "";
}
console.log(show({ b: 2, a: 1 } as { a: number }));
