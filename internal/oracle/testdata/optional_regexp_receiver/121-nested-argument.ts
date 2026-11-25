function check(outer: RegExp | undefined, inner: RegExp | undefined): boolean | undefined {
    return outer?.test(String(inner?.test("entry")));
}
console.log(String(check(/true/, /entry/)));
console.log(String(check(/undefined/, undefined)));
console.log(String(check(undefined, /entry/)));
