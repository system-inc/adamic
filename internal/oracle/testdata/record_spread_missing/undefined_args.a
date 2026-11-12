// A record spread from undefined copies nothing (#bzrdem0, s1): { ...undefined, results } holds
// results alone, as Node's does.
interface Args {
	[key: string]: number | undefined;
}
function withResults(args: Args | undefined, results: number | undefined): Args {
	return { ...args, results };
}
const merged = withResults(undefined, 3);
console.log(Object.keys(merged).join(','));
console.log(String(Object.keys(merged).length));
console.log(String(merged['results']));
const empty = withResults(undefined, undefined);
console.log(Object.keys(empty).join(','));
console.log(String(Object.keys(empty).length));
console.log(String(JSON.stringify(empty)));
