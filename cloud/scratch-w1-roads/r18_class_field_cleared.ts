class Owner {
	seen: WeakSet<object> | undefined = new WeakSet<object>();
}
const owner = new Owner();
const key = {};
owner.seen?.add(key);
owner.seen = undefined;
console.log(`${owner.seen === undefined}`);
