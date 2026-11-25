function check(regex: RegExp | undefined): string {
    return regex?.test("entry") === undefined ? "skipped" : "tested";
}
console.log(check(undefined));
console.log(check(/other/));
