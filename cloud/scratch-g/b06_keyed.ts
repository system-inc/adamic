const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const key: string = process.argv.length > 99 ? 'x' : 'year';
	console.log(match.groups[key] ?? 'none');
}
