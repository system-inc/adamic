// Object.keys of a caught Error: Node prints 0, natively its message and name were fields (#bbhwj30).
try {
	throw new Error('m');
} catch (e) {
	if (e instanceof Error) {
		console.log(`${Object.keys(e).length}`);
	}
}
