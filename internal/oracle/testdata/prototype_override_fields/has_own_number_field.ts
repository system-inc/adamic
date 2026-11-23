function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
const override = { message: "plain", hasOwnProperty: 5 };
console.log(String(probe(override)));
