Runtime handoff: closures_throw.a

Ownership: runtime only. Driver flags and emitted exception checks are correct.
Fault: internal/native/runtime/sort.c, order(). When a comparator sets adamic_thrown, its ADAMIC_TARGET_WASI branch calls adamic_panic instead of propagating the throw. sorted_or_stopped() also has its setjmp recovery compiled out; adamic_timsort() therefore cannot report a stopped sort back to the generated caller. The generated caller already checks adamic_thrown. The first nine callback exception probes agree; execution stops at the first throwing sort comparator.
The runtime worker must unwind sort's own work without losing the pending exception, preserve the original array and free sort temporaries. No runtime fix was authored here.

Exact original program, copied byte for byte:

```typescript
// Function values that throw, out of every loop the runtime runs them from: map, the visits,
// reduce, Array.from and sort, and a call written out. Each loop lets go of what it holds as the
// throw leaves it, and sort leaves the array as it was, as V8 does. The strings are built at run
// time, so a reference not let go, or let go twice, is caught.

function word(stem: string, index: number): string {
	return `${stem}${index}`;
}

function attempt(label: string, run: () => string): void {
	try {
		console.log(`${label}: ${run()}`);
	} catch (error) {
		console.log(`${label} threw: ${error instanceof Error ? error.message : 'something'}`);
	}
}

const words = [word('a', 1), word('b', 2), word('c', 3), word('d', 4)];

attempt('map', () => words.map((value, index) => {
	if (index === 2) {
		throw new Error(word('map stopped at ', index));
	}
	return `${value}!`;
}).join(','));
attempt('filter', () => words.filter((value) => {
	if (value === 'c3') {
		throw new Error(word('filter stopped at ', 3));
	}
	return value !== 'a1';
}).join(','));
attempt('find', () => words.find((value) => {
	if (value === 'b2') {
		throw new Error(`find stopped at ${value}`);
	}
	return false;
}) ?? 'none');
attempt('some', () => `${words.some((value) => {
	if (value.startsWith('d')) {
		throw new Error(`some stopped at ${value}`);
	}
	return false;
})}`);
attempt('forEach', () => {
	const seen: string[] = [];
	words.forEach((value) => {
		seen.push(value);
		if (seen.length === 3) {
			throw new Error(`forEach stopped after ${seen.join('+')}`);
		}
	});
	return seen.join('+');
});
attempt('reduce', () => words.reduce((carried, value) => {
	if (value === 'c3') {
		throw new Error(`reduce stopped with ${carried}`);
	}
	return `${carried}${value}`;
}, word('start', 0)));
attempt('Array.from', () => Array.from({ length: 5 }, (_, index) => {
	if (index === 3) {
		throw new Error(word('from stopped at ', index));
	}
	return word('made', index);
}).join(','));

const ages = new Map<string, string>();
ages.set(word('ann', 1), word('age ', 31));
ages.set(word('bob', 2), word('age ', 42));
ages.set(word('cy', 3), word('age ', 53));
attempt('Map forEach', () => {
	const seen: string[] = [];
	ages.forEach((value, key) => {
		if (key === 'bob2') {
			throw new Error(`Map forEach stopped at ${key} with ${value} after ${seen.join('+')}`);
		}
		seen.push(`${key}=${value}`);
	});
	return seen.join('+');
});
const names = new Set<string>([word('ann', 1), word('bob', 2), word('cy', 3)]);
attempt('Set forEach', () => {
	let count = 0;
	names.forEach((name) => {
		count += 1;
		if (name.startsWith('cy')) {
			throw new Error(`Set forEach stopped at ${name}`);
		}
	});
	return `${count}`;
});
console.log(`${ages.size} ${names.size}`);

// A sort whose comparator throws leaves the array as it was.
const letters = [word('q', 9), word('d', 4), word('m', 1), word('a', 7), word('z', 2), word('k', 5)];
let compared = 0;
attempt('sort', () => letters.sort((left, right) => {
	compared += 1;
	if (compared === 6) {
		throw new Error(`sort stopped after ${compared} comparisons`);
	}
	return left < right ? -1 : left > right ? 1 : 0;
}).join(','));
console.log(letters.join(','));
function strict(left: string, right: string): number {
	if (left === 'a7' || right === 'a7') {
		throw new Error('the comparator met a7');
	}
	return left < right ? -1 : 1;
}
attempt('sort by name', () => letters.sort(strict).join(','));
console.log(letters.join(','));
function sortedStrictly(list: string[]): string {
	list.sort(strict);
	return `sorted ${list.join(',')}`;
}
attempt('sort by name, in a function', () => `${sortedStrictly(letters)} and after`);
console.log(letters.join(','));
const many: string[] = [];
for (let index = 0; index < 100; index += 1) {
	many.push(word('n', (index * 37) % 100));
}
// Two runs of 50 take 441 comparisons to sort by insertion (counted on Node); the 500th is in their merge.
let manyCompared = 0;
attempt('sort, mid-merge', () => many.sort((left, right) => {
	manyCompared += 1;
	if (manyCompared === 500) {
		throw new Error('sort stopped in a merge');
	}
	return left < right ? -1 : left > right ? 1 : 0;
}).slice(0, 3).join(','));
console.log(many.slice(0, 5).join(','));
const maybes: (number | undefined)[] = [3, undefined, 1, 2];
attempt('sort with undefined', () => maybes.sort((left, right) => {
	if (left === 1 || right === 1) {
		throw new Error('the comparator met 1');
	}
	return (left ?? 0) - (right ?? 0);
}).join(','));
console.log(maybes.join(','));

// A map assigned back to the array it maps, then read by the catch: a throw leaves the array as it was.
let mapped = [word('x', 1), word('y', 2), word('z', 3)];
try {
	mapped = mapped.map((value) => {
		if (value === 'z3') {
			throw new Error('mapped in place stopped');
		}
		return `${value}${value}`;
	});
} catch (error) {
	console.log(error instanceof Error ? error.message : 'something');
}
console.log(mapped.join(','));
// Here the array is dead after the map on every path, so it could be written over in place, but
// for a callback that can throw it isn't: the throw leaves the loop the way any map's does.
function remapped(count: number): string {
	const items: string[] = [];
	for (let index = 0; index < count; index += 1) {
		items.push(word('w', index));
	}
	const doubled = items.map((value) => {
		console.log(`remapping ${value}`);
		if (value === 'w2') {
			throw new Error(`remapped stopped at ${value}`);
		}
		return `${value}${value}`;
	});
	return doubled.join(',');
}
attempt('remapped', () => remapped(2));
attempt('remapped', () => remapped(4));

// A call written out, its temporaries let go, and a throw through a function value calling another.
const failing = (): string => {
	throw new Error(word('called', 1));
};
const outer = (): string => `${word('before', 0)}${failing()}`;
attempt('call', () => `${word('left', 1)} ${outer()} ${word('right', 2)}`);
function passes(): string {
	return words.map((value) => (value === 'b2' ? failing() : value)).join('');
}
attempt('through a function', passes);
attempt('finally inside', () => {
	try {
		return failing();
	} finally {
		console.log('the closure\'s finally ran');
	}
});
```

Node source oracle: exit 0

stdout (UTF-8 JSON string):

```json
"map threw: map stopped at 2\nfilter threw: filter stopped at 3\nfind threw: find stopped at b2\nsome threw: some stopped at d4\nforEach threw: forEach stopped after a1+b2+c3\nreduce threw: reduce stopped with start0a1b2\nArray.from threw: from stopped at 3\nMap forEach threw: Map forEach stopped at bob2 with age 42 after ann1=age 31\nSet forEach threw: Set forEach stopped at cy3\n3 3\nsort threw: sort stopped after 6 comparisons\nq9,d4,m1,a7,z2,k5\nsort by name threw: the comparator met a7\nq9,d4,m1,a7,z2,k5\nsort by name, in a function threw: the comparator met a7\nq9,d4,m1,a7,z2,k5\nsort, mid-merge threw: sort stopped in a merge\nn0,n37,n74,n11,n48\nsort with undefined threw: the comparator met 1\n3,,1,2\nmapped in place stopped\nx1,y2,z3\nremapping w0\nremapping w1\nremapped: w0w0,w1w1\nremapping w0\nremapping w1\nremapping w2\nremapped threw: remapped stopped at w2\ncall threw: called1\nthrough a function threw: called1\nthe closure's finally ran\nfinally inside threw: called1\n"
```

stderr (UTF-8 JSON string):

```json
""
```

WASI with emitted-C fix and runtime 52959fc: exit 70

stdout (UTF-8 JSON string):

```json
"map threw: map stopped at 2\nfilter threw: filter stopped at 3\nfind threw: find stopped at b2\nsome threw: some stopped at d4\nforEach threw: forEach stopped after a1+b2+c3\nreduce threw: reduce stopped with start0a1b2\nArray.from threw: from stopped at 3\nMap forEach threw: Map forEach stopped at bob2 with age 42 after ann1=age 31\nSet forEach threw: Set forEach stopped at cy3\n3 3\n"
```

stderr (UTF-8 JSON string):

```json
"adamic: panic: wasm32: throwing sort comparators are not supported\n"
```

Raw source, stdout and stderr are retained in this directory; observations.json in the parent records all five fixtures before and after.
