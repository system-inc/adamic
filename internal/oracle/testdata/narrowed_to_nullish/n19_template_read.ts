let kept: number | undefined = undefined;
function keep(value: number): number { kept = value; return value; }
console.log(`${keep(4)} ${kept}`);
