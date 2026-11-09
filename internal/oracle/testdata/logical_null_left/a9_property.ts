class Item { value = 1; ready = true; }
function find(flag: boolean): Item | null { return flag ? new Item() : null; }
const o = { item: find(false) };
const value = o.item && o.item.ready;
if (value === null) { console.log('null'); } else { console.log(value ? 'true' : 'false'); }
