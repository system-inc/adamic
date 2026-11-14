const f = await Promise.resolve<(a: string) => string>((...items: string[]): string => items.join(","));
const two: (a: string, b: number) => string = f;
console.log(two("x", 2));
