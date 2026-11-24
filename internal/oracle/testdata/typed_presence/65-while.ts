function pick(index: number): { readonly code: number } | undefined {
    return index < 3 ? { code: index * 10 } : undefined;
}
function copy(source: { readonly code: number } | undefined) {
    return { ...source };
}
let index = 0;
let current = copy(pick(index));
while ("code" in current) {
    console.log(String(current.code));
    index = index + 1;
    current = copy(pick(index));
}
console.log(String(index));
