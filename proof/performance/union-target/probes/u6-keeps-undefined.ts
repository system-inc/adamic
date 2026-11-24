interface A { readonly a: number; readonly shared: string }
interface B { readonly b: string; readonly shared: string }
interface C { readonly c: boolean }
function f(x: A | B | C | undefined): A | undefined { return x as A | undefined; }
console.log(String(f({ a: 1, shared: "s" })?.a));
console.log(String(f(undefined)?.a));
