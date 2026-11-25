const sized: { readonly size: number }[] = [{ size: 1 }];
const maps: Map<string, number>[] = [new Map(), new Map()];
sized.splice(1, 0, ...maps);
console.log(String(sized.length));
