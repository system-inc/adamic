function maybe(n: number): { id: number; label: string } | undefined { return n > 0 ? undefined : { id: 1, label: 'y' }; }
const c = { ...maybe(1), label: 'x' };
console.log(c.label);
