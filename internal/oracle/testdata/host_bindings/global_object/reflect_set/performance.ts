// onProfilerEvent as tsc's performance.ts declares it and reads it in mark, Absent on every target.
declare let onProfilerEvent: ((eventName: string) => void) | undefined;

export function mark(markName: string): void {
    if (typeof onProfilerEvent === "function") {
        onProfilerEvent(markName);
    }
}
