interface A { readonly kind: "a"; readonly shared: string }
interface B { readonly kind: "b"; readonly shared: string }
interface C { readonly c: boolean }
function f(x: A | B | C): string { return (x as A | B).shared; }
console.log(f({ kind: "a", shared: "from a" }));
