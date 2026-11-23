interface Comparer {
    compare(left: string, right: string): number;
}
const comparer: Comparer = new Intl.Collator(undefined, { usage: "sort", sensitivity: "variant", numeric: true });
if (comparer.compare) {
    console.log(String(comparer.compare("a", "b")));
} else {
    console.log("no compare");
}
