// A closure writes the captured boolean | undefined each way, true, false and undefined, and another
// closure reads the cell after each write.
function run(): void {
	let seen: boolean | undefined;
	const write = (step: number): void => {
		if (step === 0) {
			seen = true;
		} else if (step === 1) {
			seen = false;
		} else {
			seen = undefined;
		}
	};
	const read = (): string => `${seen}`;
	for (let step = 0; step < 4; step++) {
		write(step % 3);
		console.log(`step ${step} ${read()}`);
	}
}

run();
