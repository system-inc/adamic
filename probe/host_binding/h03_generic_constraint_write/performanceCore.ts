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
function rewrite<T extends { now: () => number }>(clock: T): void {
	clock.now = (): number => 0;
}
export function run(): void {
	if (typeof performance === "object") {
		const clock: Clock = performance;
		rewrite(clock);
		console.log(performance.now() > 0 ? "later" : "origin");
	}
}
