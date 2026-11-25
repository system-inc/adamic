function check(regex: RegExp | undefined): string {
    return regex?.test("entry").toString() ?? "absent";
}
console.log(check(/entry/));
console.log(check(undefined));
