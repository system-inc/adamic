const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const view: Record<string, string> = match.groups;
	console.log(view.toString());
}
