class Item {
    value = 1;
}
const items: Item[] = [new Item()];
const item = items[0];
console.log(`${item && item.value === 1}`);
