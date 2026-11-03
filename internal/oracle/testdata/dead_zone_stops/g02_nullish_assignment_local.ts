function fill(): number { let local: number | undefined = undefined; local ??= 4; return local; }
console.log(String(fill()));
