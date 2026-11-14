const box = { join(...items: string[]): string { return items.map((item) => typeof item + ":" + item).join(","); } };
const one: { join(a: string): string } = box;
const two: { join(a: string, b: number): string } = one;
console.log(two.join("x", 2));
