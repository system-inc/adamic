// Closed gap 1: the return type is monomorphized to string at this call.
function identity<Item>(item: Item): Item {
	return item;
}
console.log(identity('a'));
