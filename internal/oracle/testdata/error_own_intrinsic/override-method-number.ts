function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
console.log(String(probe(new Error("x"))));
class Override {
    message = "plain";
    hasOwnProperty(_key: number): boolean { return false; }
}
console.log(String(probe(new Override())));
