// Memory's m1: one closure writes each member in turn while another reads after each write, and
// the object it ends on is then overwritten, so the counted build sees that object freed.
class Session {
	readonly name: string;
	readonly parts: number[];
	constructor(name: string) {
		this.name = name;
		this.parts = [1, 2];
	}
}

function alternate(round: number): void {
	let active: Session | "stopping" | undefined;
	const write = (step: number): void => {
		const member = step % 3;
		if (member === 0) {
			active = new Session(`session ${round}.${step}`);
		} else if (member === 1) {
			active = "stopping";
		} else {
			active = undefined;
		}
	};
	const read = (): string => {
		if (active === undefined) {
			return "undefined";
		}
		if (active === "stopping") {
			return "stopping";
		}
		return `${active.name} with ${active.parts.length}`;
	};
	const seen: string[] = [];
	for (let step = 0; step < 7; step++) {
		write(step);
		seen.push(read());
	}
	write(1);
	seen.push(read());
	console.log(seen.join(", "));
}

alternate(0);
alternate(1);
