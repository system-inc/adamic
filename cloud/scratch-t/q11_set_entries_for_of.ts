const set = new Set<number>([1, 2]);
const entries = set.entries();
console.log(entries.toString());
for (const [key, value] of entries) {
	console.log(key, value);
}
