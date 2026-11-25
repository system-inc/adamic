// An item that throws after a spread (#p1qw4kf): JavaScript evaluates every item, each spread expanded, before
// splice removes or inserts anything, so the throw leaves the array as it was, with a count of 0 or more, and
// the items packed so far, made strings among them, are released on the way to the catch.
function boom(): number {
	throw new Error('boom');
}

function word(): string {
	throw new Error('word');
}

function caught(label: string, left: string): void {
	console.log(`${label}: caught ${left}`);
}

const items = [8, 9];
const values = [0, 1, 2, 3];
try {
	values.splice(1, 0, ...items, boom());
} catch {
	caught('count 0', String(JSON.stringify(values)));
}
try {
	values.splice(1, 2, ...items, boom());
} catch {
	caught('count 2', String(JSON.stringify(values)));
}
try {
	values.splice(-1, 1, ...items, boom(), ...items);
} catch {
	caught('between two spreads', String(JSON.stringify(values)));
}
const removed = values.splice(1, 1, ...items);
console.log(`after the catches: ${String(JSON.stringify(removed))} ${String(JSON.stringify(values))}`);

// Strings that were made, so a packed item not released on the throw is a leak.
const words = ['ant', 'bee', 'cat'].map((name) => `${name}s`);
const more = ['elk', 'fox'].map((name) => name.repeat(2));
try {
	words.splice(1, 1, ...more, `${more.length}`, word());
} catch {
	caught('strings', String(JSON.stringify(words)));
}
