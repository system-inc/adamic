// An async function's captured boolean | undefined lives in a cell of its frame: written after an
// await, and read by a closure the function made before it. given, the optional parameter, is
// captured too, so it goes into its frame cell from the call.
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
