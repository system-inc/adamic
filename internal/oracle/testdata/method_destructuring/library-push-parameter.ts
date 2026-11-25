interface P { push(...items: number[]): number }
function append({ push }: P): number { return push(1); }
console.log(String(append([])));
