// onProfilerEvent as tsc's performance.ts declares it (line 15) and reads it in mark (line 80). It is
// Absent on every target, so typeof is "undefined" on Node and on every backend, the guard is false,
// and the call under it is never made.
declare let onProfilerEvent: ((eventName: string) => void) | undefined;

let marks = 0;

export function mark(markName: string): void {
    marks += 1;
    if (typeof onProfilerEvent === "function") {
        onProfilerEvent(markName);
    }
}

export function markUnlessUndefined(markName: string): void {
    marks += 1;
    if (typeof onProfilerEvent !== "undefined") onProfilerEvent(markName);
}

export function counted(): number {
    return marks;
}

export function profiler(): string {
    return typeof onProfilerEvent;
}
