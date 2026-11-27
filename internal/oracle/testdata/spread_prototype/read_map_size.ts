const built: { readonly size: number } = new Map<string, number>();
const copy = { ...built };
console.log(typeof copy.size);
