function check(regex: RegExp | undefined): string {
    return String((regex)?.test("entry"));
}
console.log(check(undefined));
console.log(check(/entry/));
