// sys.require loads a plugin from the path it resolved.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
const modulePath: string = "./plugin.js";
require(modulePath);
