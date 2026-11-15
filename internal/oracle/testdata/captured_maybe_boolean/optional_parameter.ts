// A captured optional boolean parameter, given nothing, undefined, true or false, read by the closure
// its function returns: fsWatch's recursive once it's widened to boolean | undefined.
function watcher(name: string, recursive?: boolean): () => string {
	return (): string => `${name} ${recursive} ${recursive ?? false} ${recursive === undefined}`;
}

console.log(watcher("nothing")());
console.log(watcher("undefined", undefined)());
console.log(watcher("true", true)());
console.log(watcher("false", false)());
