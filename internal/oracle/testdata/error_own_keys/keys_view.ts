// Object.keys of a caught Error through a structural view: Node prints 0, native printed 2.
try {
	throw new Error('m');
} catch (e) {
	if (e instanceof Error) {
		const o: { message: string } = e;
		console.log(`${Object.keys(o).length}`);
	}
}
