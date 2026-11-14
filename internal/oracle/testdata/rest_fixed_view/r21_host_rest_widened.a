import 'node:timers';
type Ruled = (callback: (label?: unknown) => void, ms: number, label?: unknown) => unknown;
type Wider = (callback: (label?: unknown) => void, ms: number, label?: unknown, extra?: unknown) => unknown;
const ruled: Ruled = setTimeout;
const wider: Wider = ruled;
wider((label?: unknown): void => { console.log("fired " + (typeof label === "string" ? label : "other")); }, 1, "a", "b");
