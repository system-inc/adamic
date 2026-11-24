const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const copy = { ...match.groups };
	console.log(copy['year'] ?? 'none');
}
