class Pair {
	both(name: string, count: number): void { console.log(`${name}:${String(count)}`); }
}
interface Both {
	both(...args: [name: string, count: number]): void;
}
const pair = new Pair();
const view: Both = pair;
view.both(`n${String(2)}`, 3);
