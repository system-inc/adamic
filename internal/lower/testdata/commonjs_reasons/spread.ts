// A require called with a spread keeps a reason of its own: only a path given as its one argument
// lowers, as the throw a computed require is.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
const modulePath: [string] = ["./plugin.js"];
require(...modulePath);
