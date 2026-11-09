class Item { value = 1; ready = true; }
function find(flag: boolean): Item | null { return flag ? new Item() : null; }
const none = find(false);
const value = none && none.ready;
if (value === null) { console.log('null'); } else { console.log(value ? 'true' : 'false'); }
