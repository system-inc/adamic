const sized: { readonly length: number }[] = [{ length: 1 }];
const lists: number[][] = [[1], [2, 3]];
sized.push(...lists);
console.log(String(sized.length));
