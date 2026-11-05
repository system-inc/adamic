class Item {
	label: string;
	constructor(label: string) { this.label = label; }
	read(): string { return this.label; }
}
class Holder { child: Item | undefined = undefined; }
function family(index: number): void {
	const holder = new Holder();
	holder.child = new Item(`first${index}`);
	const saved = holder.child;
	holder.child = undefined;
	console.log(`${saved.read()}/${holder.child === undefined}`);
}
for (const index of [1, 2]) { family(index); }
