// sort: a million numbers sorted with a comparator, then a million more already nearly in order,
// the two cases TimSort is built for. Array reads and writes and a callback per comparison.

let seed = 1;
function next(): number {
	seed = (seed * 1103515245 + 12345) % 2147483648;
	return seed / 2147483648;
}

const count = 1000000;
const random: number[] = [];
for (let index = 0; index < count; index += 1) {
	random.push(next() * 1000000);
}
random.sort((left, right) => left - right);

const nearly: number[] = [];
for (let index = 0; index < count; index += 1) {
	nearly.push(index + (next() < 0.01 ? next() * 1000 : 0));
}
nearly.sort((left, right) => left - right);

let checksum = 0;
for (let index = 0; index < count; index += 1000) {
	checksum += (random[index] ?? 0) + (nearly[index] ?? 0);
}
let ordered = true;
for (let index = 1; index < count; index += 1) {
	if ((random[index - 1] ?? 0) > (random[index] ?? 0) || (nearly[index - 1] ?? 0) > (nearly[index] ?? 0)) {
		ordered = false;
	}
}
console.log(`${ordered} ${checksum.toFixed(3)}`);
