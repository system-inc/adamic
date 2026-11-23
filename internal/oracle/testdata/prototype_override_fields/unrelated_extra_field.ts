function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
const extra = { message: "plain", count: 1 };
console.log(String(probe(extra)));
