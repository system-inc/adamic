const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const held: { readonly [key: string]: string | undefined } = match.groups;
	console.log(JSON.stringify(held) ?? 'none');
}
