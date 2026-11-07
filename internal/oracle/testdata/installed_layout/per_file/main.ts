// The ruled per-file-module-directory divergence (#redxes2, Adamic's ruling on #0ta3dk9): each module's own
// __dirname, printed raw. Node, run a file per module, gives each its own directory, lib/util.ts's a level below this
// one's; one bundled image gives both the image's directory.
import { utilDirectory } from './lib/util.ts';

console.log(`util ${utilDirectory}`);
console.log(`main ${__dirname}`);
