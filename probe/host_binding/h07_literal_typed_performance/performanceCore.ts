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
	const own: Performance = { timeOrigin: 0, now: (): number => 5, mark(): void {}, measure(): void {}, clearMarks(): void {}, clearMeasures(): void {} };
	console.log(String(own.now()));
}
