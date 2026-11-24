const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	try {
		console.log(match.groups.toString());
	} catch (error) {
		if (error instanceof Error) {
			console.log(error.message);
		}
	}
	console.log('after');
}
