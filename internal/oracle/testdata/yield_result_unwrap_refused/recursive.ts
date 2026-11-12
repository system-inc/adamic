// Recursive S retains the same walk state; compile must finish, not unfold forever.
interface S { readonly nested: IterableIterator<S> | undefined; readonly m: ReadonlyMap<string, number>; }
type Items = IterableIterator<S>;
function consume(items: Items): number {
 for (const item of items) return item.m.size;
 return 0;
}
function run() {
const elem: S = { nested: undefined, m: new Map<string, number>([['a', 1]]) };
return consume((function* () { if (elem) yield elem; })());
}
console.log(String(run()));
