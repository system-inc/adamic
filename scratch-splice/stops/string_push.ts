const sized: { readonly length: number }[] = [{ length: 1 }];
const names = ['ab', 'cde'];
sized.push(...names);
console.log(String(sized.length));
