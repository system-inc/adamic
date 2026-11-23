const nums: { k: number; [key: string]: number } = { k: 41 };
nums['j'] = 1;
const seen: { readonly k: number } = nums;
console.log(String(seen.k.toFixed(1)));
