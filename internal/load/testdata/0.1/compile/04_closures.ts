function makeCounter(start: number): () => number {
	let count = start;
	return () => {
		count += 1;
		return count;
	};
}

const first = makeCounter(0);
const second = makeCounter(100);
first();
first();
console.log(`first: ${first()}, second: ${second()}`);

// Each iteration gets its own binding of `step`, as in JavaScript.
const adders: ((value: number) => number)[] = [];
for (let step = 1; step <= 3; step += 1) {
	adders.push((value) => value + step);
}
console.log(adders.map((add) => add(10)).join(' '));
