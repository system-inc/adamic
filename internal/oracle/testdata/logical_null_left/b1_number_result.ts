class Item { value = 1; ready = true; }
function find(flag: boolean): Item | null { return flag ? new Item() : null; }
const none = find(false);
const value = none && none.value;
console.log(String(value === null));
const some = find(true);
const other = some && some.value;
console.log(String(other === null));
