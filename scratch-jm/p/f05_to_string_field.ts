function probe(value: { message: string }): string {
    return value.toString();
}
const override = { message: "plain", toString: (): string => "mine" };
console.log(probe(override));
