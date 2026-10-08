interface View { value: number; }
class State { value: number = undefined!; }
function read(state: View): void { console.log(`${state.value}`); }
read(new State());
