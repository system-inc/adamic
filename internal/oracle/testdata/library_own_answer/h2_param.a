type Groups = NonNullable<RegExpMatchArray['groups']>;
function show(groups: Groups): void {
	console.log(groups.toString());
}
show({ year: '2024' });
const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	show(match.groups);
}
