type First = { readonly kind: 'a'; readonly name: string };
type Second = { readonly kind: 'b'; readonly name: string };
type Third = { readonly kind: 'c'; readonly name: string };
type Item = First | Second | Third;
interface Holder { item: Item; self: Holder | undefined }
function narrow(value: Item): First | Second { return value as First | Second; }
function turn(index: number): string {
	const holder: Holder = { item: { kind: 'a', name: 'a' + String(index) }, self: undefined };
	holder.self = holder;
	return narrow(holder.item).name;
}
let last = '';
for (let index = 0; index < 1000; index++) last = turn(index);
console.log(last);
