class Item { value = 1; ready = true; }
function maybe(flag: boolean): Item | undefined { return flag ? new Item() : undefined; }
const none = maybe(false);
const value = none && none.ready;
console.log(String(value === undefined));
console.log(String(value === null));
