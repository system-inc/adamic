const named: { readonly name: string }[] = [{ name: 'a' }];
const callbacks: (() => number)[] = [() => 1, () => 2];
named.push(...callbacks);
console.log(String(named.length));
