function copy(source: { readonly code?: number }) {
    return { ...source };
}
function absentCopy(source: { readonly code: number } | undefined) {
    return { ...source };
}
function has(value: { readonly code?: number }): boolean {
    return "code" in value;
}
console.log(String(has(copy({}))));
console.log(String(has(copy({ code: 1 }))));
console.log(String(has(absentCopy(undefined))));
