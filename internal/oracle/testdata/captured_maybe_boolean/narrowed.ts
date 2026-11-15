// Closures narrow the captured boolean | undefined and read it as the boolean it's narrowed to,
// while another closure writes it.
function run(): void {
	let seen: boolean | undefined;
	const describe = (): string => {
		if (seen === undefined) {
			return "unset";
		}
		return seen ? "on" : "off";
	};
	const negated = (): string => (seen !== undefined && !seen ? "off for sure" : "not off");
	const write = (value: boolean | undefined): void => {
		seen = value;
	};
	console.log(`${describe()}, ${negated()}`);
	write(true);
	console.log(`${describe()}, ${negated()}`);
	write(false);
	console.log(`${describe()}, ${negated()}`);
	write(undefined);
	console.log(`${describe()}, ${negated()}`);
}

run();
