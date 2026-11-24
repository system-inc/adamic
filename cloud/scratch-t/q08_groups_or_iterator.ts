function show(value: NonNullable<RegExpMatchArray['groups']> | MapIterator<string>): void {
	console.log(value.toString());
}
const map = new Map<string, number>([['a', 1]]);
show(map.keys());
const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	show(match.groups);
}
