// Node's answers for each library object: toString, toLocaleString, Object.prototype.toString.call, a template, String.
function* generator(): Generator<number> { yield 1; }
async function* asyncGenerator(): AsyncGenerator<number> { yield 1; }
const map = new Map<string, number>([['a', 1]]);
const set = new Set<number>([1]);
const rows: [string, () => unknown][] = [
	['MapIterator entries', () => map.entries()],
	['MapIterator keys', () => map.keys()],
	['MapIterator values', () => map.values()],
	['MapIterator iterator', () => map[Symbol.iterator]()],
	['SetIterator values', () => set.values()],
	['SetIterator entries', () => set.entries()],
	['ArrayIterator values', () => [1].values()],
	['ArrayIterator entries', () => [1].entries()],
	['ArrayIterator keys', () => [1].keys()],
	['TypedArray values', () => new Uint8Array(1).values()],
	['StringIterator', () => 'ab'[Symbol.iterator]()],
	['RegExpStringIterator', () => 'a1'.matchAll(/[0-9]/g)],
	['Generator', () => generator()],
	['AsyncGenerator', () => asyncGenerator()],
	['Promise', () => Promise.resolve(1)],
	['WeakMap', () => new WeakMap<object, number>()],
	['WeakSet', () => new WeakSet<object>()],
	['WeakRef', () => new WeakRef({})],
	['Map', () => map],
	['Set', () => set],
	['Date', () => new Date(0)],
	['RegExp', () => /a/g],
	['Error', () => new Error('e')],
	['TypeError', () => new TypeError('e')],
	['Uint8Array', () => new Uint8Array(1)],
	['ArrayBuffer', () => new ArrayBuffer(1)],
	['DataView', () => new DataView(new ArrayBuffer(1))],
	['Intl.NumberFormat', () => new Intl.NumberFormat('en-US')],
	['Intl.Collator', () => new Intl.Collator('en-US')],
	['TextEncoder', () => new TextEncoder()],
	['JSON', () => JSON],
	['Math', () => Math],
	['iterator result', () => map.entries().next()],
	['RegExp match', () => 'a1'.match(/[0-9]/)],
	['RegExp exec', () => /[0-9]/.exec('a1')],
	['groups', () => '2024'.match(/(?<year>[0-9]+)/)?.groups],
	['object literal', () => ({ x: 1 })],
];
function attempt(f: () => string): string {
	try {
		return f();
	} catch (error) {
		return error instanceof Error ? error.name + ': ' + error.message : 'thrown';
	}
}
for (const [name, make] of rows) {
	const x = make() as { toString(): string; toLocaleString(): string };
	console.log(name);
	console.log('  toString       ' + attempt(() => x.toString()));
	console.log('  toLocaleString ' + attempt(() => x.toLocaleString()));
	console.log('  O.p.toString   ' + attempt(() => Object.prototype.toString.call(x)));
	console.log('  template       ' + attempt(() => `${x}`));
	console.log('  String         ' + attempt(() => String(x)));
	console.log('  toStringTag    ' + attempt(() => String((x as { [Symbol.toStringTag]?: unknown })[Symbol.toStringTag])));
}
