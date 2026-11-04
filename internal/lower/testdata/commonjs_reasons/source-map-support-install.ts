// sys.ts calls install on what the require gives, and the require throws before the call is made.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
require("source-map-support").install();
