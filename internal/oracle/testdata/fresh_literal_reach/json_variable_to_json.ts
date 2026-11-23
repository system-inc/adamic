function show(value: { a: number }): string {
    return JSON.stringify(value) ?? "";
}
const made = { a: 1, toJSON: (): string => "custom" };
console.log(show(made));
