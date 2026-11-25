const regex = /a/g;
function pattern(enabled: boolean): RegExp | undefined { return enabled ? regex : undefined; }
function check(): void {
    console.log(String(pattern(true)?.test("aa")));
    console.log(String(regex.lastIndex));
}
check();
check();
