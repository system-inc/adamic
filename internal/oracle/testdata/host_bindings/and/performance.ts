// A read of onProfilerEvent on the right of &&, whose left is its typeof guard: as a statement, and as
// a boolean value, which is the left's false on Node and on every backend.
declare let onProfilerEvent: ((eventName: string) => void) | undefined;

export function fire(name: string): boolean {
    typeof onProfilerEvent === "function" && onProfilerEvent(name);
    return typeof onProfilerEvent === "function" && onProfilerEvent.length > 0;
}
