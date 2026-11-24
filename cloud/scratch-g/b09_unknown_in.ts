const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const held: unknown = match.groups;
	if (typeof held === 'object' && held !== null) {
		console.log(String('year' in held));
	}
}
