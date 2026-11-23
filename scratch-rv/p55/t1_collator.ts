interface Comparer {
    compare(left: string, right: string): number;
}
const comparer: Comparer = new Intl.Collator("en");
if (comparer.compare) {
    console.log("has compare");
} else {
    console.log("no compare");
}
