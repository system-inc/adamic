const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	console.log('before');
	console.log(String(match.groups.propertyIsEnumerable('year')));
}
