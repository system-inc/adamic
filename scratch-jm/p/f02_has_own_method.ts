function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
class Override {
    message = "plain";
    hasOwnProperty(_key: number): boolean { return false; }
}
console.log(String(probe(new Override())));
