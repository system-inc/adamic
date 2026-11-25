// An empty spread inserts nothing and still removes (#p1qw4kf); a splice with no count, or no spread, lowers as
// it did before.
function show(label: string, removed: number[], left: number[]): void {
	console.log(`${label}: [${removed.join(',')}] [${left.join(',')}]`);
}

const none: number[] = [];
let values = [0, 1, 2, 3];
show('empty array', values.splice(1, 2, ...none), values);
values = [0, 1, 2, 3];
show('empty literal', values.splice(1, 1, ...[]), values);
values = [0, 1, 2, 3];
show('empty with an item', values.splice(1, 1, 7, ...none), values);
values = [0, 1, 2, 3];
show('no count', values.splice(1), values);
values = [0, 1, 2, 3];
show('no spread', values.splice(1, 1, 5, 6), values);
