// Each value the captured cell held is let go of when it's overwritten: an object by "stopping",
// "stopping" by undefined, an object by another object, a built string by an object, and the last
// value when the cell itself goes. The leak-checked builds say whether every count came back.
class Session {
	readonly name: string;
	readonly parts: number[];
	constructor(name: string) {
		this.name = name;
		this.parts = [1, 2, 3];
	}
}

function cycle(round: number): string {
	let active: Session | "stopping" | undefined;
	const write = (step: number): void => {
		if (step === 0) {
			active = new Session(`first ${round}`);
		} else if (step === 1) {
			active = "stopping";
		} else if (step === 2) {
			active = undefined;
		} else if (step === 3) {
			active = new Session(`second ${round}`);
		} else if (step === 4) {
			active = new Session(`third ${round}`);
		} else if (step === 5) {
			const word = ["stop", "ping"].join("");
			if (word === "stopping") {
				active = word;
			}
		} else {
			active = new Session(`last ${round}`);
		}
	};
	const see = (): string => {
		if (active === undefined) {
			return "undefined";
		}
		if (active === "stopping") {
			return "stopping";
		}
		return `${active.name} of ${active.parts.length}`;
	};
	const seen: string[] = [];
	for (let step = 0; step < 7; step++) {
		write(step);
		seen.push(see());
	}
	return seen.join(", ");
}

for (let round = 0; round < 3; round++) {
	console.log(cycle(round));
}
