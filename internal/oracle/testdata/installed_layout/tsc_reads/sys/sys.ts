// tsc's reads of its own path, held in a module of their own as tsc's sys.ts holds them, a directory away from the
// entry that calls them: the directory it runs from, the default lib beside it (getExecutingFilePath and
// getDefaultLibFolder), the case-sensitivity probe of sys.ts:1728, and a locale's messages as utilitiesPublic.ts:721
// loads them under --locale ja. Each answers what tsc derives, never the path itself. In one bundled image this
// module's __filename and __dirname are the image's, as in _tsc.js (Adamic's ruling on #0ta3dk9).
import { existsSync, readFileSync } from 'node:fs';
import { basename, join } from 'node:path';

// sys.ts's own swapCase: every word character with its case flipped.
function swapCase(s: string): string {
    return s.replace(/\w/g, ch => {
        const up = ch.toUpperCase();
        return ch === up ? ch.toLowerCase() : up;
    });
}

export function executingDirectoryName(): string {
    return basename(__dirname);
}

export function libraryExists(name: string): boolean {
    return existsSync(join(__dirname, name));
}

export function swappedCaseExists(): boolean {
    return existsSync(swapCase(__filename));
}

export function localeMessages(locale: string): string {
    const path = join(__dirname, locale, 'diagnosticMessages.generated.json');
    return existsSync(path) ? readFileSync(path, 'utf8') : 'missing';
}
