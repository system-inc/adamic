function f(x: number): 1;
function f(x: string): 2;
function f(x: number | string): 1 | 2 { return 1; }
function consume(g: typeof f): void {}
f(1);
consume(f);
(f)(1);
const escaped = f;
namespace N {
 export function f(x: number): 1;
 export function f(x: string): 2;
 export function f(x: number | string): 1 | 2 { return 1; }
}
N.f(1);
const member = N.f;
const shorthand = { f };
