interface Box<T> { make(): T }
class Maker<T> { constructor(readonly v: T) {} make(): T { return this.v; } }
const box: Box<number> = new Maker(3);
const { make } = box;
console.log(String(make()));
