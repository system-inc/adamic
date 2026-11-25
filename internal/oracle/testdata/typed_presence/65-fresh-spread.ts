function has(value: { code: number } | undefined): boolean {
    return "code" in { ...value };
}
console.log(String(has(undefined)));
console.log(String(has({ code: 1 })));
