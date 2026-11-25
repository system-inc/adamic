// A spread whose elements have another representation from the array's, functions into objects that have a
// name, stops as push stops it (#p1qw4kf).
const named: { readonly name: string }[] = [{ name: 'a' }];
const callbacks: (() => number)[] = [() => 1, () => 2];
named.splice(1, 0, ...callbacks);
console.log(String(named.length));
