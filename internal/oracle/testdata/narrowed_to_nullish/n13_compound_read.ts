function fill(): void { s = 'x'; }
let s: string | undefined = undefined;
fill();
s += 'a';
console.log(String(s));
