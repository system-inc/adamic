// The onProfilerEvent row's declaration, as tsc's performance.ts writes it, and its typeof.
declare let onProfilerEvent: ((eventName: string) => void) | undefined;

console.log(typeof onProfilerEvent);
