let calls = 0;
// Unproven, and it lies on false: it says no to an array.
function isList(value: unknown): value is any[] {
    calls = calls + 1;
    return calls > 1000;
}
// f has Array.isArray's type, but its symbol is f, so its answer is isList's.
const f: typeof Array.isArray = isList;
function g(x: number[] | string): string {
    if (!f(x)) {
        return String(x.length);
    }
    return '';
}
console.log(g([1, 2]));
