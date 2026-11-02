// "#x" in an instance whose class has a private x: false on Node, and natively in judges a class
// instance by its public shape, so it agrees (#a1b746w's twin of message_hash).
class Box {
	#x = 1;
	read(): number {
		return this.#x;
	}
}
const box = new Box();
console.log(`${box.read()} ${'#x' in box}`);
