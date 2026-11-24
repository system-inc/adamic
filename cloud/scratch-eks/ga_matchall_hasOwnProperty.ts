for (const found of '2024'.matchAll(/(?<year>[0-9]+)/g)) {
	if (found.groups !== undefined) {
		console.log('before');
		console.log(String(found.groups.hasOwnProperty('year')));
	}
}
