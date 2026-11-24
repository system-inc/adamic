const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const held: unknown = match.groups;
	const key = ['x', 'year'][1] ?? 'x';
	const view = held as { [key: string]: unknown };
	console.log(String(view[key]));
}
