class State {
	readonly definitions = new Map<number, number>();
}
const states = new Map<number, State>();
const state = new State();
state.definitions.set(1, 7);
states.set(0, state);
const found = states.get(0)?.definitions.get(1);
const missing = states.get(9)?.definitions.get(1);
console.log(`${found} ${missing}`);
