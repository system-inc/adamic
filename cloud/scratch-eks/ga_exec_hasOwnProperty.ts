const found = /(?<year>[0-9]+)/.exec('2024');
if (found !== null && found.groups !== undefined) {
	console.log('before');
	console.log(String(found.groups.hasOwnProperty('year')));
}
