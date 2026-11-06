// A read of onProfilerEvent in the true arm of ?:, whose test is its typeof guard: the false arm is the
// value, a number and a string.
declare let onProfilerEvent: ((eventName: string) => void) | undefined;

export function arity(): number {
    return typeof onProfilerEvent !== "undefined" ? onProfilerEvent.length : -1;
}

export function named(): string {
    return typeof onProfilerEvent === "function" ? onProfilerEvent.name : "absent";
}
