interface Tester {
    test(text: string): boolean;
}
const tester: Tester = /a/;
if (tester.test) {
    console.log(String(tester.test("cat")));
} else {
    console.log("no test");
}
