interface Tester {
    test(text: string): boolean;
}
const tester: Tester = /a/;
console.log(tester.test ? "has test" : "no test");
