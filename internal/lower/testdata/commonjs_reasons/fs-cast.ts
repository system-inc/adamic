// A cast of a counted require, of fs, is no Modelled form, so its any is judged: the cast still stops.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
interface FileSystem {
	existsSync(path: string): boolean;
}
console.log(String((require("fs") as FileSystem).existsSync("./plugin.js")));
