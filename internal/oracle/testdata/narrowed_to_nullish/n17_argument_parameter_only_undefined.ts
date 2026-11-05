let kept: number | undefined = undefined;
function keep(value: number): number { kept = value; return value; }
function show(value: undefined): string { return String(value); }
console.log(`${keep(4)} ${show(kept)}`);
