class Source {
    get y(): number { return 7; }
}
const view: { readonly y: number } = new Source();
const copy = { ...view };
console.log(String("y" in copy));
