function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
const plain = { message: "plain" };
console.log(String(probe(plain)));
