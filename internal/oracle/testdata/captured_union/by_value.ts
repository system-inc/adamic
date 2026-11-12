// The literal member is compared by value: a "stopping" the program builds at run time is a
// different string from the interned constant, and === must still say they're equal, whichever
// of the two the captured cell holds.
class Session {
	readonly name: string;
	constructor(name: string) {
		this.name = name;
	}
}

function built(): string {
	return ["stop", "ping"].join("");
}

function compare(): void {
	let active: Session | "stopping" | undefined;
	const word = built();
	const write = (step: number): void => {
		if (step === 0) {
			active = new Session("first");
		} else if (step === 1) {
			active = "stopping";
		} else if (step === 2) {
			active = undefined;
		} else if (word === "stopping") {
			// A heap string the checker narrowed to the literal, held in the cell.
			active = word;
		}
	};
	const matches = (): string => `${active === word} ${word === active} ${active === built()} ${active === "stopping"} ${active !== "stopping"}`;
	for (let step = 0; step < 4; step++) {
		write(step);
		console.log(`step ${step} ${matches()}`);
	}
	const named = (): string => {
		const fresh = built();
		if (fresh === "stopping" && active === fresh) {
			return "the cell holds the word";
		}
		return "the cell holds something else";
	};
	console.log(named());
	write(0);
	console.log(named());
	write(1);
	console.log(named());
}
compare();
