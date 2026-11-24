const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const key = (): string => {
		console.log('key');
		return 'year';
	};
	console.log(String(match.groups.hasOwnProperty(key())));
}
