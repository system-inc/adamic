interface Chain {
    readonly text: string;
    readonly code?: number;
}
function copy(source: Chain | undefined) {
    return { ...source };
}
function full<T extends Chain>(value: T): boolean {
    return "text" in value && "code" in value;
}
function named<T extends { readonly text?: string }>(value: T): boolean {
    return "text" in value;
}
console.log(String(full({ text: "a", code: 1 })));
console.log(String(full({ text: "b" })));
console.log(String(named(copy(undefined))));
console.log(String(named(copy({ text: "c" }))));
