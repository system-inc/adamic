function show(value: { a: number }): string {
    return JSON.stringify(value) ?? "";
}
console.log(show({ a: 1, toJSON: (): string => "custom" } as { a: number }));
