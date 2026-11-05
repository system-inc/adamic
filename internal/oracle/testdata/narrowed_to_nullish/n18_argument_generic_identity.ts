let kept: number | undefined = undefined;
function keep(value: number): number { kept = value; return value; }
function id<T>(value: T): T { return value; }
console.log(`${keep(4)} ${String(id(kept))}`);
