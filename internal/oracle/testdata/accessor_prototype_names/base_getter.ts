// A base class's getter named after an Object.prototype member, read through a subclass, calls the getter on
// both backends, as Node does (#76y1mth). Node prints before, 1.
class Base {
	get valueOf(): number {
		return 1;
	}
}
class Leaf extends Base {
	tag = 'leaf';
}
console.log('before');
const leaf = new Leaf();
console.log(String(leaf.valueOf));
