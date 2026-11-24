interface Stamp {
	toISOString(): string;
}
const stamp: Stamp = new Date(0);
console.log(stamp.toString());
