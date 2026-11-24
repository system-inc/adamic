const map = new Map<string, number>([['a', 1]]);
const holder = { iterator: map.values() };
console.log(holder.iterator.toString());
console.log(Object.prototype.toString.call(holder.iterator));
