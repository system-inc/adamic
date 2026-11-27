interface Switch {
	readonly label: string;
	readonly open: boolean;
}
class Gate {
	readonly label = "g";
	get open(): boolean {
		return true;
	}
}
const built: Switch = new Gate();
const copy: Switch = { ...built };
console.log(`${built.open} ${copy.open}`);
console.log(typeof copy.open + " " + String(copy.open === undefined));
