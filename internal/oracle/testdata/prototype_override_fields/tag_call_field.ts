function probe(value: { message: string }): string {
    return Object.prototype.toString.call(value);
}
const override = { message: "plain", toString: (): string => "mine" };
console.log(probe(override));
