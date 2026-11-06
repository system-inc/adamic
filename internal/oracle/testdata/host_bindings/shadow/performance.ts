// A parameter named onProfilerEvent shadows the declared one inside run, so its guard and its call are
// the parameter's, matched by symbol, never by name: the call is made, and typeof says "function". The
// declared name, read in declared, is still the Absent host binding.
declare let onProfilerEvent: ((eventName: string) => void) | undefined;

export function run(onProfilerEvent: (eventName: string) => void, name: string): string {
    if (typeof onProfilerEvent === "function") {
        onProfilerEvent(name);
    }
    return typeof onProfilerEvent;
}

export function declared(): string {
    return typeof onProfilerEvent;
}
