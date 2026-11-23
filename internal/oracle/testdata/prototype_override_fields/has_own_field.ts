function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
const override = { message: "plain", hasOwnProperty: (_key: number): boolean => false };
console.log(String(probe(override)));
