function check(regex: RegExp | undefined): void {
    const result = regex?.test("entry");
    console.log(String(result));
}
check(/entry/);
check(/other/);
check(undefined);
