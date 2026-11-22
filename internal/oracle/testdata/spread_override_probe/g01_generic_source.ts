function over<T extends object>(s: T): string { const c = { ...s, label: 'x' }; return c.label; }
console.log(over({ id: 1 }));
