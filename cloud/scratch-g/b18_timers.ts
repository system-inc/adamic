import 'node:timers';
function viewed(value: unknown): unknown {
	return value;
}
const match = 'x'.match(/(?<_onTimeout>x)/);
if (match !== null && match.groups !== undefined) {
	clearTimeout(viewed(match.groups));
	console.log('cleared');
	console.log(String(match.groups['_onTimeout']));
}
