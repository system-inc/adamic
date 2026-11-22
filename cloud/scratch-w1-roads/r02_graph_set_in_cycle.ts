interface Cell {
	peer: Cell | undefined;
	seen: WeakSet<Cell>;
}
const first: Cell = { peer: undefined, seen: new WeakSet<Cell>() };
const second: Cell = { peer: first, seen: new WeakSet<Cell>() };
first.peer = second;
first.seen.add(second);
console.log(`${first.seen.has(second)} ${second.seen.has(first)}`);
