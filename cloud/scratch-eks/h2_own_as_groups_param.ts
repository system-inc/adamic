type Groups = NonNullable<RegExpMatchArray['groups']>;
function show(groups: Groups): string {
	return groups.toString();
}
const match = '2024'.match(/(?<year>[0-9]+)/);
console.log(show({ year: '1999' }));
if (match !== null && match.groups !== undefined) {
	console.log(show(match.groups));
}
