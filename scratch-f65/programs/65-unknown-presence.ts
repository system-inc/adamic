function copy(source: { readonly code: number } | undefined) {
    return { ...source };
}
function hasCode(value: unknown): boolean {
    return typeof value === "object" && value !== null && "code" in value;
}
console.log(String(hasCode(copy(undefined))));
console.log(String(hasCode(copy({ code: 7 }))));
