function input(): string { console.log("evaluated"); return "entry"; }
function check(pattern: RegExp | null | undefined): void {
    console.log(String(pattern?.test(input())));
}
check(null);
check(undefined);
check(/entry/);
