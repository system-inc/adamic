// An Error made with a message, read the same way as no_message.a: its own message is found.
function msg(e: unknown): string {
	return typeof e === 'object' && e !== null && 'message' in e && typeof e.message === 'string' ? e.message : 'none';
}
try {
	throw new Error('x');
} catch (e) {
	console.log(JSON.stringify(msg(e)) ?? 'undefined');
}
