function probe(value: { message: string }): boolean {
    return value.propertyIsEnumerable("message");
}
const override = { message: "plain", propertyIsEnumerable: (_key: number): boolean => false };
console.log(String(probe(override)));
