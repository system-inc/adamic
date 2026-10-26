let calls = 0;
// Unproven, and it lies on false: it says no to an array.
function isList(value: unknown): value is any[] {
    calls = calls + 1;
    return calls > 1000;
}
// A mapped field names Array.isArray's declaration, but it's a symbol of its own, holding isList.
const fake: Pick<ArrayConstructor, 'isArray'> = { isArray: isList };
function g(x: number[] | string): string {
    if (!fake.isArray(x)) {
        return String(x.length);
    }
    return '';
}
console.log(g([1, 2]));
