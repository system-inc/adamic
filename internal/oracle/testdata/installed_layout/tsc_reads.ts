// What tsc derives from its own path where it's installed, one line per read, and never the path itself, which
// differs by the program file's own name (_tsc.js on Node, _tsc natively), a declared divergence (#redxes2). Each
// read is tsc's: the directory it runs from, the default lib beside it (sys.ts's getExecutingFilePath and
// getDefaultLibFolder), the case-sensitivity probe of sys.ts:1728, and a locale's messages as
// utilitiesPublic.ts:721 loads them under --locale ja. It's built only as a CommonJS program, as _tsc.js ships.
import { existsSync, readFileSync } from 'node:fs';
import { basename, join } from 'node:path';

// sys.ts's own swapCase: every word character with its case flipped.
function swapCase(s: string): string {
    return s.replace(/\w/g, ch => {
        const up = ch.toUpperCase();
        return ch === up ? ch.toLowerCase() : up;
    });
}

const locale = join(__dirname, 'ja', 'diagnosticMessages.generated.json');
console.log(`directory ${basename(__dirname)}`);
console.log(`lib.d.ts ${existsSync(join(__dirname, 'lib.d.ts'))}`);
console.log(`swapped case exists ${existsSync(swapCase(__filename))}`);
console.log(`locale ${existsSync(locale) ? readFileSync(locale, 'utf8') : 'missing'}`);
console.log(`lib.es2020.d.ts ${existsSync(join(__dirname, 'lib.es2020.d.ts'))}`);
