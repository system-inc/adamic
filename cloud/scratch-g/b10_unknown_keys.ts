const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const held: object = match.groups;
	console.log(Object.keys(held).join(','));
}
