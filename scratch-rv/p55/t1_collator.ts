interface Comparer {
    compare(left: string, right: string): number;
}
const comparer: Comparer = new Intl.Collator("en");
if (comparer.compare) {
    console.log(String(comparer.compare("a", "b")));
} else {
    console.log("no compare");
}
