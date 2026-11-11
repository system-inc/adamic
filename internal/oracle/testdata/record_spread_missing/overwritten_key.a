// A record spread from a present Args | undefined copies its entries (#bzrdem0, s2): a key the
// literal writes again keeps the spread's position and takes the later value, as Node's does, and
// the spread record is copied, not changed.
interface Args {
	[key: string]: number | undefined;
}
function withResults(args: Args | undefined, results: number | undefined): Args {
	return { ...args, results };
}
const held: Args = { first: 1, second: 2, results: 9 };
const merged = withResults(held, 3);
console.log(Object.keys(merged).join(','));
console.log(String(Object.keys(merged).length));
console.log(String(merged['results']));
const middle: Args = { results: 9, last: 2 };
const moved = withResults(middle, 4);
console.log(Object.keys(moved).join(','));
console.log(String(Object.keys(moved).length));
console.log(String(JSON.stringify(moved)));
console.log(String(JSON.stringify(held)));
console.log(String(JSON.stringify(middle)));
