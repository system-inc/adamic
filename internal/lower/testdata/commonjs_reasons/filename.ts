// sys.ts reads its own path from __filename.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
const executingFilePath: string = __filename;
console.log(executingFilePath);
