// #tq6yzqy's Record trace, the twin main already admits: the same road with an optional number field,
// which recordStorageType has held in an intersection since 05ddb9a12. Whatever this does, the optional
// boolean road does too, so it says whether the road is new with #tq6yzqy or older. Node prints: 1, then
// undefined, then 2.
type Counts = Partial<Record<"count", number>>;
type Entry = { count?: number } & { count?: number };
function probe(value: Entry): void {
	console.log(String(value.count));
}
const set: Counts = { count: 1 };
const unset: Counts = {};
probe(set);
probe(unset);
const direct: Entry = { count: 2 };
probe(direct);
