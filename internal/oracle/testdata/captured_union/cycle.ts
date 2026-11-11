// Memory's m3: an object member reached from the cell holds a closure that captures the cell, a
// cycle through the Union: the Session the cell holds keeps report, and report keeps the cell.
class Session {
	readonly name: string;
	readonly report: () => string;
	constructor(name: string, report: () => string) {
		this.name = name;
		this.report = report;
	}
}

function run(round: number): string {
	let active: Session | "stopping" | undefined;
	const report = (): string => {
		if (active === undefined) {
			return "undefined";
		}
		if (active === "stopping") {
			return "stopping";
		}
		return `session ${active.name}`;
	};
	const seen: string[] = [report()];
	active = new Session(`first ${round}`, report);
	seen.push(active.report());
	active = "stopping";
	seen.push(report());
	active = new Session(`second ${round}`, report);
	seen.push(active.report());
	active = undefined;
	seen.push(report());
	active = new Session(`last ${round}`, report);
	seen.push(report());
	return seen.join(", ");
}

console.log(run(0));
console.log(run(1));
