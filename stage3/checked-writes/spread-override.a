const source: { count: number } = { count: 7 };
const narrow: { readonly count: 1 | 2 } = { ...source, count: 1 };
const view: { count: number } = narrow;
view.count = 3;
