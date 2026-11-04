// A counted require, of fs, returned as an object literal's member is no Modelled form, so its any is
// judged: the literal's relation to the typed slot is recorded, and the require keeps the generic
// reason it stops at today. Node's globals are typed only for a program that imports a node: module.
import 'node:process';
interface Loaded {
	m: {};
}
function load(): Loaded {
	return { m: require("fs") };
}
load();
