const narrow: { readonly count: 0 | 1 } = { count: 0 };
const view: { count: number } = narrow;
view.count += 1;
console.log(narrow.count.toString());
