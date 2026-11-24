interface A { readonly a: number; readonly shared: string }
interface B { readonly b: string; readonly shared: number }
interface C { readonly c: boolean }
function f(x: A | B | C): string { return String((x as A | B).shared); }
console.log(f({ a: 1, shared: "from a" }));
