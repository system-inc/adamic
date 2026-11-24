const pattern = ['(?<year>[0-9]+)'][0] ?? '';
const match = '2024'.match(new RegExp(pattern));
if (match !== null && match.groups !== undefined) {
	console.log('before');
	console.log(match.groups.toString());
}
