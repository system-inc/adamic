declare const performance: Performance | undefined;
interface Performance {
	readonly timeOrigin: number;
	now(): number;
	mark(name: string): void;
	measure(name: string, startMark?: string, endMark?: string): void;
	clearMarks(name?: string): void;
	clearMeasures(name?: string): void;
}
interface Clock { now(): number; }
function read(source: Performance): number { return source.now(); }
export function run(): void {
	const own: Performance = { timeOrigin: 0, now: (): number => 5, mark(): void {}, measure(): void {}, clearMarks(): void {}, clearMeasures(): void {} };
	console.log(String(read(own)));
	if (typeof performance === "object") { console.log(read(performance) > 0 ? "host" : "zero"); }
}
