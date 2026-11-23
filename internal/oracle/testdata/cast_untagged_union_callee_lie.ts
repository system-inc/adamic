interface X { readonly a: number; readonly b: string }
interface Y { readonly b: number; readonly a: string }
interface A { readonly child: X; readonly f: number }
interface B { readonly f: number; readonly g: string; readonly child: Y }
// The lie read in a callee compiled for A (#axqgdck, Language's y04): the view's checks travel with the value, so the
// read of a.child.b stops there, after the honest call agrees. Node runs on and prints number.
function make(which: number): A | B {
    const a: A = { child: { a: 1, b: 'x' }, f: 3 };
    const b: B = { f: 6, g: 'g', child: { b: 4, a: 'z' } };
    return which === 0 ? a : b;
}
function readChild(a: A): string {
    return typeof a.child.b;
}
console.log(readChild(make(0) as A));
console.log(readChild(make(1) as A));
