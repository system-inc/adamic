// "#x" in an instance whose class has a private x: false on Node, but the JavaScript backend keeps
// privates as # string keys (#a1b746w).
class Box {
	#x = 1;
	read(): number {
		return this.#x;
	}
}
const box = new Box();
console.log(`${box.read()} ${'#x' in box}`);
