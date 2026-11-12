// This module's own IteratorYieldResult gets the ordinary walk: fail closed.
export {};
interface IteratorYieldResult<T> { readonly value: T; }
interface Node { readonly m: ReadonlyMap<string, number>; }
interface Child extends Node { readonly tag: number; }
type Source = { readonly result: IteratorYieldResult<Child> };
type Target = { readonly result: IteratorYieldResult<Node> };
type Items = IterableIterator<Target>;
function consume(items: Items): number {
 for (const item of items) return item.result.value.m.size;
 return 0;
}
const child: Child = { tag: 1, m: new Map<string, number>([['a', 1]]) };
function run() {
const elem: Source = { result: { value: child } };
return consume((function* (): Generator<Source, void, unknown> { if (elem) yield elem; })());
}
console.log(String(run()));
