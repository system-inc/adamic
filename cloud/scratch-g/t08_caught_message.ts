const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	try {
		match.groups.valueOf();
	} catch (error) {
		if (error instanceof TypeError) {
			console.log(error.message);
		}
	}
}
