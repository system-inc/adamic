let best: number | undefined;
const pick = (): number => { best = 1; return 0; };
best = undefined;
console.log(`${pick()} ${best === undefined}`);
