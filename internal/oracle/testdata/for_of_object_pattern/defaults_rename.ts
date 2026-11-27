// A default and a renamed binding in a for...of object pattern, as a declaration lowers them (#rm22wzv).
type Pair = { count?: number; name: string };
const pairs: Pair[] = [{ name: "first" }, { count: 5, name: "second" }];
let total = 0;
for (const { count = 1, name: label } of pairs) {
    total += count;
    console.log(label + " " + String(count));
}
console.log(String(total));
