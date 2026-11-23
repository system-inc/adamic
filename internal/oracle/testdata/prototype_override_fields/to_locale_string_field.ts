function probe(value: { message: string }): string {
    return value.toLocaleString();
}
const override = { message: "plain", toLocaleString: (): string => "mine" };
console.log(probe(override));
