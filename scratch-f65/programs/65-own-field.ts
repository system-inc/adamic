function stamped(source: { readonly code: number; readonly name: string } | undefined) {
    return { ...source, code: 5 };
}
function show(value: { readonly code?: number; readonly name?: string }): string {
    return String("code" in value) + " " + String("name" in value) + " " + String(value.code);
}
console.log(show(stamped(undefined)));
console.log(show(stamped({ code: 1, name: "one" })));
