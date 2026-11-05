interface Item { text: string; }
interface Holder { value: Item | null; }
function consume(value: Item | null): string { return value === null ? 'null' : value.text; }
function observe(holder: Holder): void {
	if (holder.value === null) console.log(consume(holder.value));
}
const holder: Holder = { value: { text: 'item' } };
holder.value = null;
observe(holder);
