const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	console.log(Object.entries(match.groups).map(([key, value]) => key + '=' + String(value)).join(','));
}
