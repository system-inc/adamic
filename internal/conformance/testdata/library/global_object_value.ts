// Spec: the global object as a value (an alias, an argument, a spread, `valueOf()`) is not built yet
// Expect: notyet the global object as a value
// On Node, Reflect.set installs a global through the global object handed out by valueOf, and the
// guarded call prints "set mark". The stop keeps any lowering of the value from reopening that road.
const fire = (eventName: string): void => {
	console.log(`set ${eventName}`);
};
Reflect.set(globalThis.valueOf(), 'onProfilerEvent', fire);
console.log('done');
