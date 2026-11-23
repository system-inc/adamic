function probe(value: { message: string }): boolean {
    return value.isPrototypeOf({});
}
const override = { message: "plain", isPrototypeOf: (_other: object): boolean => true };
console.log(String(probe(override)));
