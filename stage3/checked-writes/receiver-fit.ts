const original: { readonly count: 1 | 2 } = { count: 1 };
const replacement: { count: number } = { count: 7 };
let view: { count: number } = original;
function value(): number { view = replacement; return 2; }
view.count = value();
console.log(original.count.toString() + ' ' + replacement.count.toString());
