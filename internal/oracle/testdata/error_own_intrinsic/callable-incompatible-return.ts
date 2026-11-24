function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
console.log(String(probe(new Error("x"))));
const override = { message: "plain", hasOwnProperty: (_key: string): number => 7 };
console.log(String(probe(override)));
