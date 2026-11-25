// The array spread into itself: JavaScript expands the spread before splice removes or inserts anything, so the
// items are the array as it was (#p1qw4kf).
function show(label: string, removed: number[], left: number[]): void {
	console.log(`${label}: [${removed.join(',')}] [${left.join(',')}]`);
}

let values = [1, 2, 3];
show('inserted', values.splice(1, 0, ...values), values);
values = [1, 2, 3];
show('removing first', values.splice(0, 2, ...values), values);
values = [1, 2, 3];
show('removing after', values.splice(1, 2, ...values), values);

const words = ['ant', 'bee'].map((word) => `${word}s`);
const taken = words.splice(1, 1, ...words);
console.log(`${taken.join(',')} | ${words.join(',')}`);
