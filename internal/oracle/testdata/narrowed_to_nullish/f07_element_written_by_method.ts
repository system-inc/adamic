const t: (number | undefined)[] = [1];
function f(): void { t.fill(4); }
t[0] = undefined;
f();
console.log(String(t[0]));
