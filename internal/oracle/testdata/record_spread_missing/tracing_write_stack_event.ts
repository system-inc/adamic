// tracing.ts's writeStackEvent cut down to its spread (#bzrdem0, s3): { ...args, results } at
// tracing.ts:183, args read from an event whose args is optional, Args declared as tsc declares it.
// writeEvent prints what tsc writes to its trace file, without the file. The event is read through a
// guard, since Adamic checks with noUncheckedIndexedAccess, which tsc's own build doesn't.
// writeEvent's JSON.stringify(args) stops lowering (a later wall, #d9ep7de), so this program is pinned
// as that stop. writeStackEvent is declared first: function bodies lower in declaration order, so the
// stop in writeEvent is reached only once the spread has lowered, where main stops at the spread.
interface Args {
	[key: string]: string | number | boolean | null | undefined | Args | readonly (string | number | boolean | null | undefined | Args)[];
}
const eventStack: { phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean }[] = [];
const sampleInterval = 1000 * 10;
function writeStackEvent(index: number, endTime: number, results?: Args): void {
	const event = eventStack[index];
	if (event === undefined) return;
	const { phase, name, args, time, separateBeginAndEnd } = event;
	if (separateBeginAndEnd) {
		writeEvent("E", phase, name, args, /*extras*/ undefined, endTime);
	}
	else if (sampleInterval - (time % sampleInterval) <= endTime - time) {
		writeEvent("X", phase, name, { ...args, results }, `"dur":${endTime - time}`, time);
	}
}
function writeEvent(eventType: string, phase: string, name: string, args: Args | undefined, extras?: string, time: number = 0): void {
	let line = `{"pid":1,"tid":1,"ph":"${eventType}","cat":"${phase}","ts":${time},"name":"${name}"`;
	if (extras) line += `,${extras}`;
	if (args) {
		line += `,"args":${JSON.stringify(args)}`;
		line += ` keys ${Object.keys(args).join(',')} count ${Object.keys(args).length}`;
	}
	console.log(`${line}}`);
}
eventStack.push({ phase: "check", name: "checkSourceFile", time: 0, separateBeginAndEnd: false });
writeStackEvent(0, 20000, { count: 2, nested: { deep: true }, list: ["a", 1, null] });
eventStack.push({ phase: "emit", name: "emitJsFileOrBundle", args: { path: "x.ts", results: "old" }, time: 5, separateBeginAndEnd: false });
writeStackEvent(1, 20000);
eventStack.push({ phase: "bind", name: "bindSourceFile", args: { path: "y.ts" }, time: 7, separateBeginAndEnd: false });
writeStackEvent(2, 20000, { kind: 1 });
eventStack.push({ phase: "parse", name: "createSourceFile", time: 9, separateBeginAndEnd: false });
writeStackEvent(3, 20000);
eventStack.push({ phase: "program", name: "createProgram", args: { configFilePath: "tsconfig.json" }, time: 11, separateBeginAndEnd: true });
writeStackEvent(4, 20000);
