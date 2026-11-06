// Language's p11 on #00qn5az: Reflect.set installs onProfilerEvent on the global object before mark reads
// it, so on Node typeof turns "function" and the guarded call prints, where a native fold would print
// nothing. Adamic stops it as the global object as a value.
import { mark } from './performance.ts';

Reflect.set(globalThis, "onProfilerEvent", (eventName: string): void => {
    console.log(`set ${eventName}`);
});
mark('mark');
