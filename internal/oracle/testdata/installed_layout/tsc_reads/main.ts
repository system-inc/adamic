// What tsc derives from its own path where it's installed, one line per read, and never the path itself, which
// differs by the program file's own name (_tsc.js on Node, _tsc natively), a declared divergence (#redxes2). The reads
// are sys/sys.ts's, a module a directory away, so the program is built as one bundled image on both sides: one
// _tsc.js on Node, as tsc ships, and one image natively, where every module's __filename and __dirname are the
// image's (Adamic's ruling on #0ta3dk9).
import { executingDirectoryName, libraryExists, localeMessages, swappedCaseExists } from './sys/sys.ts';

console.log(`directory ${executingDirectoryName()}`);
console.log(`lib.d.ts ${libraryExists('lib.d.ts')}`);
console.log(`swapped case exists ${swappedCaseExists()}`);
console.log(`locale ${localeMessages('ja')}`);
console.log(`lib.es2020.d.ts ${libraryExists('lib.es2020.d.ts')}`);
