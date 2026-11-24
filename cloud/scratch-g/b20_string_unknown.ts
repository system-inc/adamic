const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const held: unknown = match.groups;
	console.log('before');
	console.log(String(held));
}
