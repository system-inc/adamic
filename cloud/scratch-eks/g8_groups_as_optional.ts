const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const view: { year?: string } = match.groups;
	console.log(view.toString());
}
