const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	console.log(match.groups['year'] ?? 'none');
	console.log(match.groups.year ?? 'none');
}
