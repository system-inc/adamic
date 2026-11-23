function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
console.log(String(probe(new Error("x"))));
const override = { message: "plain", hasOwnProperty: (_key: PropertyKey): boolean => false };
console.log(String(probe(override)));
