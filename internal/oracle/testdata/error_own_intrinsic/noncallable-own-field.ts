function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
console.log(String(probe(new Error("x"))));
const shadow = { message: "plain", hasOwnProperty: 1 };
console.log(String(probe(shadow)));
