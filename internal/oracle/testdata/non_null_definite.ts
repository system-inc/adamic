function run(): void {
    let number!: number;
    let text!: string;
    number = 0;
    text = ''.repeat(2);
    console.log(`${number} ${text.length}`);
}
class State {
    count!: number;
    text!: string;
    constructor() { this.count = 0; this.text = ''.repeat(2); }
    read(): number { return this.count; }
}
run();
const state = new State();
console.log(`${state.read()} ${state.text.length}`);
