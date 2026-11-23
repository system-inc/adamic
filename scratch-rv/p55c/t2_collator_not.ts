interface Comparer {
    compare(left: string, right: string): number;
}
function pick(comparer: Comparer): string {
    return !comparer.compare ? "no compare" : String(comparer.compare("b", "a"));
}
console.log(pick(new Intl.Collator(undefined, { usage: "sort", sensitivity: "variant", numeric: true })));
console.log(pick({ compare: (left: string, right: string): number => left.length - right.length }));
