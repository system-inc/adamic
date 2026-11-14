// The function that declares the captured boolean | undefined reads it after each call to the
// closure that writes it: in a template, through ??, against undefined, negated and as a condition.
function run(start: boolean | undefined): void {
	let seen: boolean | undefined = start;
	const write = (step: number): void => {
		if (step === 0) {
			seen = true;
		} else if (step === 1) {
			seen = false;
		} else {
			seen = undefined;
		}
	};
	console.log(`start ${seen} ${seen ?? true} ${seen === undefined} ${!seen} ${seen ? "on" : "off"}`);
	for (let step = 0; step < 3; step++) {
		write(step);
		console.log(`step ${step} ${seen} ${seen ?? true} ${seen === undefined} ${!seen} ${seen ? "on" : "off"}`);
	}
}

run(undefined);
run(false);
