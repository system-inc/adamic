// A member typed any beside a Modelled require's is a program local's any, which the literal's other
// member doesn't excuse: the return still stops, at that member.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
interface Loaded {
	module: {};
	name: string;
}
function requirePlugin(modulePath: string): Loaded {
	const loose: any = modulePath;
	return { module: require(modulePath), name: loose };
}
requirePlugin("./plugin.js");
