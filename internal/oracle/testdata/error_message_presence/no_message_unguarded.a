// no_message.a without the typeof test: the message read off unknown is converted as it is. Node prints
// "", and natively, on main, it printed "none".
function msg(e: unknown): string {
	return typeof e === 'object' && e !== null && 'message' in e ? String(e.message) : 'none';
}
try {
	throw new Error();
} catch (e) {
	console.log(JSON.stringify(msg(e)) ?? 'undefined');
}
