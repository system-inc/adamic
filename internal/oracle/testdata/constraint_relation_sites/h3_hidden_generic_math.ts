interface Options { readonly count?: number; readonly other?: string; }
function bump<T extends Options>(options: T): number { return (options.count ?? 0) + 1; }
const wide = { other: "o", count: "7" };
const narrow: { readonly other: string } = wide;
console.log(String(bump(narrow)));
