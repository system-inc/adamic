// Memory's m2: the captured Union in a graph program. The object literal's methods capture a cell
// that holds the literal itself, a cycle, so their frame joins a graph region, and the Union's cell,
// captured by the same methods, is a graph cell.
class Session {
	readonly name: string;
	constructor(name: string) {
		this.name = name;
	}
}

type Controls = { write(step: number): void; read(): string };

function run(): void {
	let active: Session | "stopping" | undefined;
	let holder: Controls | undefined;
	const controls = {
		write(step: number): void {
			if (step === 0) {
				active = new Session("first");
			} else if (step === 1) {
				active = "stopping";
			} else if (step === 2) {
				active = undefined;
			} else {
				active = new Session(`again ${step}`);
			}
		},
		read(): string {
			const held = holder === undefined ? "loose" : "held";
			if (active === undefined) {
				return `${held} undefined`;
			}
			if (active === "stopping") {
				return `${held} stopping`;
			}
			return `${held} ${active.name}`;
		},
	};
	holder = controls;
	const seen: string[] = [controls.read()];
	for (let step = 0; step < 5; step++) {
		controls.write(step);
		seen.push(controls.read());
	}
	controls.write(1);
	seen.push(controls.read());
	console.log(seen.join(", "));
}

run();
run();
