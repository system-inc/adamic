// splice with a spread among its items resolves its start as splice does before anything is removed, and
// clamps its count to what's there (#p1qw4kf). Each row prints what was removed and the array after.
function show(label: string, removed: number[], left: number[]): void {
	console.log(`${label}: [${removed.join(',')}] [${left.join(',')}]`);
}

const items = [8, 9];
let values = [0, 1, 2, 3];
show('negative start', values.splice(-1, 0, ...items), values);
values = [0, 1, 2, 3];
show('negative start removing', values.splice(-3, 2, ...items), values);
values = [0, 1, 2, 3];
show('far negative start', values.splice(-99, 1, ...items), values);
values = [0, 1, 2, 3];
show('start past the end', values.splice(99, 1, ...items), values);
values = [0, 1, 2, 3];
show('start at the end', values.splice(4, 0, ...items), values);
values = [0, 1, 2, 3];
show('count past the end', values.splice(2, 99, ...items), values);
values = [0, 1, 2, 3];
show('count over the items', values.splice(1, 3, ...items), values);
values = [0, 1, 2, 3];
show('negative count', values.splice(1, -2, ...items), values);
values = [0, 1, 2, 3];
show('NaN start and count', values.splice(0 / 0, 0 / 0, ...items), values);
values = [0, 1, 2, 3];
show('fractions', values.splice(1.9, 1.9, ...items), values);
values = [0, 1, 2, 3];
show('negative fraction', values.splice(-1.5, 1, ...items), values);
values = [0, 1, 2, 3];
show('infinite start', values.splice(1 / 0, 0, ...items), values);
values = [0, 1, 2, 3];
show('negative infinite start', values.splice(-1 / 0, 1, ...items), values);

// Strings that were made, so a reference moved wrong is a use-after-free or a leak.
const words = ['ant', 'bee', 'cat', 'dog'].map((word) => `${word}s`);
const more = ['elk', 'fox'].map((word) => word.repeat(2));
const taken = words.splice(-2, 99, ...more);
console.log(`${taken.join(',')} | ${words.join(',')}`);
