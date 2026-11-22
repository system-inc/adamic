type U = { id: number } | { id: number; label: string };
function pick(n: number): U { return n > 0 ? { id: 1 } : { id: 2, label: 'y' }; }
const c = { ...pick(1), label: 'x' };
console.log(c.label);
