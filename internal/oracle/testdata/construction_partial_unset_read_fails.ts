// construction_partial_object.ts's sideways construction, then a read of parent, a required object field the
// literal doesn't write. Node reads undefined there and runs on; Adamic stops at the read itself, since the slot
// holds no Owner (the placeholder-use contract, #dm8exc2).
interface Owner {
	name: string;
}
interface Item {
	label: string;
}
interface Big {
	name: string;
	items: Item[];
	count: number;
	parent: Owner;
	origin?: Owner;
}
function make(name: string, x: Item, origin: Owner): Big {
	return { name, items: [x], origin } as Big;
}
const partial = make('partial', { label: 'first' }, { name: 'boss' });
console.log(partial.name);
const parent: Owner = partial.parent;
console.log(parent === undefined ? 'no parent' : parent.name);
