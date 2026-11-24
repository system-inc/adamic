function copy(source: { readonly code: number } | undefined) {
    return { ...source };
}
function hasCode(value: { readonly code?: number }): boolean {
    return "code" in value;
}
console.log(String(hasCode(copy(undefined))));
console.log(String(hasCode(copy({ code: 7 }))));
