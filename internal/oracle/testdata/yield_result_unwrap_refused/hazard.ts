// Node prints true. FakeMap is held as ir.Object, not the ir.Map the target reads.
const backing = new Map<string, number>([['a', 1]]);
class FakeMap implements ReadonlyMap<string, number> {
 readonly size = 1;
 get(key: string): number | undefined { return backing.get(key); }
 has(key: string): boolean { return key === 'a'; }
 forEach(): void {}
 keys() { return backing.keys(); }
 values() { return backing.values(); }
 entries() { return backing.entries(); }
 [Symbol.iterator]() { return backing.entries(); }
}
type Source = { readonly m: FakeMap };
type Target = { readonly m: ReadonlyMap<string, number> };
type Items = IterableIterator<Target>;
function consume(items: Items): boolean {
 for (const item of items) return item.m.has('a');
 return false;
}
function run() {
const elem: Source = { m: new FakeMap() };
return consume((function* (): Generator<Source, void, unknown> { if (elem) yield elem; })());
}
console.log(String(run()));
