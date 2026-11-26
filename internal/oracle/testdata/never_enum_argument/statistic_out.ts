// executeCommandLine.ts:1285: a field read s.type of a parameter, read in the default of switch (s.type)
// and passed to assertNever.
enum StatisticType {
	time,
	count,
	memory,
}
interface Statistic {
	name: string;
	value: number;
	type: StatisticType;
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function statisticValue(s: Statistic) {
	switch (s.type) {
		case StatisticType.count:
			return "" + s.value;
		case StatisticType.time:
			return (s.value / 1000).toFixed(2) + "s";
		case StatisticType.memory:
			return Math.round(s.value / 1000) + "K";
		default:
			assertNever(s.type);
	}
}
function outside(n: number): StatisticType {
	return n as StatisticType;
}
const statistics: Statistic[] = [
	{ name: "Files", value: 12, type: StatisticType.count },
	{ name: "Check time", value: 1234, type: StatisticType.time },
	{ name: "Memory used", value: 56789, type: outside(2) },
];
for (const s of statistics) {
	console.log(s.name + ": " + statisticValue(s));
}
console.log(statisticValue({ name: "Odd", value: 1, type: outside(7) }));
