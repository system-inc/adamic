class Tally {
	total: number;
	constructor() {
		this.total = 0;
	}
	count(step: number): number {
		this.total += step;
		return this.total;
	}
}

function counted(tally: Tally | undefined): string {
	return `${tally?.count(2)}`;
}
const tally = new Tally();
console.log(counted(undefined));
console.log(counted(tally));
console.log(counted(tally));
