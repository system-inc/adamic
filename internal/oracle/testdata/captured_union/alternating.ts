// Memory's m1: one closure writes each member in turn while another reads after each write, and
// the object it ends on is then overwritten, so the counted build sees that object freed. Kept
// small, since the admission gate compiles it natively under a 10 s limit.
class Session {
	readonly name: string;
	constructor(name: string) {
		this.name = name;
	}
}

function alternate(): void {
	let active: Session | "stopping" | undefined;
	const write = (step: number): void => {
		if (step === 0) {
			active = new Session("first");
		} else if (step === 1) {
			active = "stopping";
		} else if (step === 2) {
			active = undefined;
		} else {
			active = new Session("last");
		}
	};
	const read = (): string => {
		if (active === undefined) {
			return "undefined";
		}
		if (active === "stopping") {
			return active;
		}
		return active.name;
	};
	for (let step = 0; step < 4; step++) {
		write(step);
		console.log(read());
	}
	write(1);
	console.log(read());
}

alternate();
