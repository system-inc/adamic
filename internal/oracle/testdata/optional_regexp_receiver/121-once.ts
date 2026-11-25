let count = 0;
function pattern(enabled: boolean): RegExp | undefined {
    count += 1;
    return enabled ? /entry/ : undefined;
}
console.log(String(pattern(false)?.test("entry")));
console.log(String(pattern(true)?.test("entry")));
console.log(String(count));
