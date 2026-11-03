// sys.ts reads the directory it runs from as __dirname.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
const executingDirectory: string = __dirname;
console.log(executingDirectory);
