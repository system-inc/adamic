interface Box<T> { make(): T }
function boxOf<T>(v: T): Box<T> { return { make: () => v }; }
const box = boxOf(3);
const { make } = box;
console.log(String(make()));
