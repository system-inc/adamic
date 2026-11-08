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
export function run(): void {
	if (typeof performance === "object") {
		const clock: Clock = performance;
		const time: { now: () => number } = clock;
		time.now = (): number => 0;
		console.log(performance.now() > 0 ? "later" : "origin");
	}
}
