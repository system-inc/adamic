// The lie: a C cast as A | B, union target (a)'s checked view with no test at the cast (#b5w3ycg). It stops at the
// first read of a field the value lacks, with the view's own message. Node runs on and prints undefined.
interface A { readonly a: number; readonly shared: string }
interface B { readonly b: string; readonly shared: string }
interface C { readonly c: boolean }
function f(x: A | B | C): string { return (x as A | B).shared; }
console.log("casting");
console.log(String(f({ c: true })));
