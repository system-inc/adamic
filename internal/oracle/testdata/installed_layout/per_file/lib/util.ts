// Language's fixture for the per-file difference (#0ta3dk9): this module's own __dirname, a directory below the
// entry's. Run a file per module on Node, it's this file's directory; in one bundled image it's the image's.
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';

export const utilDirectory: string = __dirname;
