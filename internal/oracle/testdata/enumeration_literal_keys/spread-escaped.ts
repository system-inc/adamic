import "./support/iterator.ts";
function inject(target: { [key: string]: number }): void {
    target["#"] = 2;
}
const source = { first: 1 };
inject(source);
console.log(Object.keys({ ...source }).join(","));
