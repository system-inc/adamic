function probe(value: { message: string }): object {
    return value.valueOf();
}
const override = { message: "plain", valueOf: (): object => ({ other: 1 }) };
console.log(JSON.stringify(probe(override)));
