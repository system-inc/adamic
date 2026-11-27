// A fixed [Phase, Level] of two numeric enums, switched over entry[1], so the read is judged by Level, the
// element's own type, and never by Phase | Level. Each enum has a value the other lacks: Phase's 5 and
// Level's 7, so a reading by the union would hold 5 a member where the switch over Level runs its default.
enum Phase {
	Parse = 0,
	Bind = 1,
	Check = 5,
}
enum Level {
	Low = 0,
	High = 1,
	Top = 7,
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function levelName(entry: [Phase, Level]): string {
	switch (entry[1]) {
		case Level.Low:
			return "low";
		case Level.High:
			return "high";
		case Level.Top:
			return "top";
		default:
			return assertNever(entry[1]);
	}
}
function outside(n: number): Level {
	return n as Level;
}
console.log(levelName([Phase.Check, Level.Top]) + " " + levelName([Phase.Parse, Level.Low]) + " " + levelName([Phase.Bind, outside(1)]) + " " + levelName([Phase.Check, outside(7)]));
console.log(levelName([Phase.Parse, outside(5)]));
