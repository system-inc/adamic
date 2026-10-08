function locals(flag: boolean): void {
    let count: number = undefined!;
    let text: string = null!;
    let yes: boolean = undefined!;
    if (flag) { count = 0; } else { count = 4; }
    text = ''.repeat(2);
    yes = false;
    console.log(`${count} ${text.length} ${yes}`);
    let captured: number = undefined!;
    const write = (): void => { captured = 23; console.log(captured.toString()); };
    const read = (): number => captured;
    write();
    console.log(read().toString());
    let loop: number = null!;
    do { loop = 8; } while (false);
    console.log(loop.toString());
    let tried: number = undefined!;
    try { tried = 9; } catch { tried = 10; }
    console.log(tried.toString());
}
class State {
    count: number = undefined!;
    text: string = null!;
    yes: boolean = undefined!;
    constructor() { this.count = 0; this.text = ''.repeat(2); this.yes = false; }
}
function defaulted(value: number = undefined!): number { value = 12; return value; }
function supplied(value: boolean = null!): boolean { return value; }
locals(true);
const state = new State();
console.log(`${state.count} ${state.text.length} ${state.yes}`);
console.log(`${defaulted()} ${defaulted(0)} ${supplied(false)}`);
class OptionalState { value?: number = undefined!; }
const optional = new OptionalState();
optional.value = 0;
console.log(`${optional.value ?? -1}`);
const absent: { value?: number } = {};
console.log(`${absent.value ?? -1}`);
const spread = { ...state, count: 5 };
console.log(`${spread.count} ${spread.text.length} ${spread.yes}`);
const reads: (() => number)[] = [];
for (let value: number = undefined!, index = 0; index < 2; index++) {
    value = index;
    reads.push(() => value);
}
console.log(`${reads[0]!()} ${reads[1]!()}`);
let mixed: string | number = undefined!;
mixed = ''.repeat(2);
console.log(`${mixed}`);
let maybe: number | undefined = undefined!;
maybe = undefined;
console.log(`${maybe ?? -1}`);
let reference: State | undefined = null!;
reference = undefined;
console.log(`${reference === undefined}`);
let maybeBoolean: boolean | undefined = undefined!;
maybeBoolean = false;
console.log(`${maybeBoolean}`);
function shadow(undefined: number): number { const value: number = undefined!; return value; }
console.log(shadow(17).toString());
