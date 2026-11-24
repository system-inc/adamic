const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const values: unknown[] = [match.groups];
	console.log(JSON.stringify(values) ?? 'none');
}
