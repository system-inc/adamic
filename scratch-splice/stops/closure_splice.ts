const named: { readonly name: string }[] = [{ name: 'a' }];
const callbacks: (() => number)[] = [() => 1, () => 2];
named.splice(1, 0, ...callbacks);
console.log(String(named.length));
