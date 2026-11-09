class Box { #secret = 1; read(): string { return String(this.#secret); } }
const box = new Box();
const copy: Box = Object.assign({}, box);
console.log(copy.read());
