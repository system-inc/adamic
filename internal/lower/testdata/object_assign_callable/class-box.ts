// Logging-only adaptation for Adamic's console.log(message: string).
class Box {
 #secret = 1;
 read() { return this.#secret; }
}
const copy: Box = Object.assign({}, new Box());
console.log("" + copy.read());
