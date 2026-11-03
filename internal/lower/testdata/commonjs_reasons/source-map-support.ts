// sys.ts installs source-map-support by require, under --enableSourceMapSupport.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
require("source-map-support");
