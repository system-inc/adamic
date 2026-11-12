// A record spread from null copies nothing, as from undefined (#bzrdem0): { ...null, results } holds
// results alone, and a present record's entries come first, as Node's do.
interface Args {
	[key: string]: number | undefined;
}
function withResults(args: Args | null, results: number | undefined): Args {
	return { ...args, results };
}
const missing = withResults(null, 5);
console.log(Object.keys(missing).join(','));
console.log(String(Object.keys(missing).length));
console.log(String(JSON.stringify(missing)));
const held: Args = { first: 1, results: 9, last: 2 };
const merged = withResults(held, 6);
console.log(Object.keys(merged).join(','));
console.log(String(Object.keys(merged).length));
console.log(String(JSON.stringify(merged)));
