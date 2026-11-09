class Item { value = 1; ready = true; }
function find(flag: boolean): Item | null { return flag ? new Item() : null; }
const none = find(false);
const value = none && none.ready;
console.log(String(value === null));
