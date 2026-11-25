// Plain items and several spreads, of numbers, strings and objects, each in its place (#p1qw4kf), and the
// call's operands evaluated in JavaScript's order: the receiver, the start, the count, then the items, each
// spread expanded before the next item runs.
interface Point {
	readonly x: number;
}

function show(label: string, removed: number[], left: number[]): void {
	console.log(`${label}: [${removed.join(',')}] [${left.join(',')}]`);
}

const first = [10, 11];
const second = [20];
const values = [0, 1, 2, 3];
show('numbers', values.splice(1, 1, 5, ...first, ...second), values);

const words = ['a', 'b', 'c'].map((word) => word.toUpperCase());
const taken = words.splice(1, 1, 'x', ...['y', 'z'], ...words, 'w');
console.log(`${taken.join(',')} | ${words.join(',')}`);

const points: Point[] = [{ x: 1 }, { x: 2 }];
const more: Point[] = [{ x: 3 }, { x: 4 }];
const shared = { x: 5 };
const removed = points.splice(1, 1, shared, ...more, ...[shared]);
console.log(`${removed.map((point) => point.x).join(',')} | ${points.map((point) => point.x).join(',')}`);
console.log(String(points[1] === shared && points[4] === shared));

const log: string[] = [];
const target = [0, 1, 2];
const source = [7, 8];
function receiver(): number[] {
	log.push('receiver');
	return target;
}
function at(value: number, label: string): number {
	log.push(label);
	return value;
}
function take(): number {
	log.push(`take ${source.length}`);
	return source.pop() ?? -1;
}
function spread(): number[] {
	log.push('spread');
	return source;
}
show('order', receiver().splice(at(1, 'start'), at(1, 'count'), ...spread(), take(), ...spread()), target);
console.log(log.join(','));
