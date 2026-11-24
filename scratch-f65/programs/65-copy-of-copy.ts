function copy(source: { readonly code: number } | undefined) {
    return { ...source };
}
const empty = copy(undefined);
const again = { ...empty };
const full = { ...copy({ code: 4 }) };
console.log(String("code" in again));
console.log(String("code" in full));
console.log(String(again.code) + " " + String(full.code));
