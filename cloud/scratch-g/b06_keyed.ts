const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const keys = ['x', 'year'];
	const key = keys[1] ?? 'x';
	console.log(match.groups[key] ?? 'none');
}
