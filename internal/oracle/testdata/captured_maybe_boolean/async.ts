// A stop: an async function's boolean | undefined, written after an await and read by a closure made
// before it. flow refuses any boolean | undefined local of an async function, captured or not, before
// its frame's cells are laid out (async_normalize.go), so given, a parameter, stops it as well.
async function pause(): Promise<void> {
	await Promise.resolve();
}

async function settle(given?: boolean): Promise<string> {
	let seen: boolean | undefined;
	const show = (): string => `${seen} ${given}`;
	console.log(`before ${show()}`);
	await pause();
	seen = given;
	console.log(`after ${show()}`);
	await pause();
	seen = undefined;
	return `last ${show()}`;
}

console.log(await settle(true));
console.log(await settle(false));
console.log(await settle());
