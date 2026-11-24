interface A { readonly a: number; readonly shared: string }
interface B { readonly b: string; readonly shared: string }
interface C { readonly c: boolean }
function f(x: A | B | C): string { return (x as A | B).shared; }
console.log("casting");
console.log(String(f({ c: true })));
