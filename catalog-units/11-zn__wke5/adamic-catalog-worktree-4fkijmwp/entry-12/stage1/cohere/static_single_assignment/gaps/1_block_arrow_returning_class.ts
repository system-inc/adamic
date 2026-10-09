class Box {
	readonly id: number;
	constructor(id: number) {
		this.id = id;
	}
}
const make = (box: Box): Box => {
	return new Box(box.id + 1);
};
console.log(`${make(new Box(1)).id}`);
