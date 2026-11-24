type Fits = { message: string; hasOwnProperty: (key: PropertyKey) => boolean };
type Wrong = { message: string; hasOwnProperty: (key: PropertyKey) => number };
function probe(value: { message: string }): boolean {
    return value.hasOwnProperty("message");
}
function choose(flag: boolean): Fits | Wrong {
    const fits: Fits = { message: "fit", hasOwnProperty: (_key: PropertyKey): boolean => false };
    const wrong: Wrong = { message: "wrong", hasOwnProperty: (_key: PropertyKey): number => 7 };
    return flag ? fits : wrong;
}
console.log(String(probe(new Error("x"))));
console.log(String(probe(choose(false))));
