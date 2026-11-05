let best: number | undefined;
const shape = { toString(): string { best = 1; return 'shape'; } };
best = undefined;
console.log(`${shape} ${best === undefined}`);
