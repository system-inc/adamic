function show(value: { a: number }): string {
    return JSON.stringify(value);
}
const extra = { toJSON: (): string => "custom" };
console.log(show({ ...extra, a: 1 }));
