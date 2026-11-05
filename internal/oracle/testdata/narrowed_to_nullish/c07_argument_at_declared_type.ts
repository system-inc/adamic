let kept: number | undefined = undefined;
function keep(value: number): number { kept = value; return value; }
function show(value: number | undefined): string { return value === undefined ? 'none' : `${value}`; }
console.log(`${keep(4)} ${show(kept)}`);
