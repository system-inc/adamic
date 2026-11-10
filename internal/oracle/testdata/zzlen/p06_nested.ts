function outer(base: number): Function {
    function inner(a: number, b: number): number { return a + b + base; }
    return inner;
}
console.log(String(outer(1).length));
