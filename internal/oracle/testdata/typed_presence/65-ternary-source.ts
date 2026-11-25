interface Coded {
    readonly code: number;
}
function choose(flag: boolean, present: Coded, absent: Coded | undefined) {
    return { ...(flag ? present : absent) };
}
function has(value: { readonly code?: number }): boolean {
    return "code" in value;
}
const present: Coded = { code: 5 };
console.log(String(has(choose(true, present, undefined))));
console.log(String(has(choose(false, present, undefined))));
console.log(String(has(choose(false, present, { code: 6 }))));
