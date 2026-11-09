class Item { value = 1; ready = true; }
function find(flag: boolean): Item | null { return flag ? new Item() : null; }
const some = find(true);
const value = some && some.ready;
if (value === null) { console.log('null'); } else { console.log(value ? 'true' : 'false'); }
