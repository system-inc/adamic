/*
 * Ask the real Tailwind engine what every shape in the corpus is, and emit the answers as JSON.
 *
 * The Go port of `inferDataType` is checked against these answers rather than against a reading of
 * the TypeScript source, because the two are not the same claim. Reading the source says `serif` is
 * a generic-name; asking the engine says it is a family-name, because `family-name` comes earlier in
 * the type list and matches first. Every expectation in this corpus is a measurement.
 *
 * The engine is reached through the bundled chunk rather than through source, because the published
 * package ships no `src/`. Bundling mangled the names, so the export is located by identity: the
 * chunk's `d` export is the function that returns null for a `var()` input and 'color' for '#fff',
 * which no other export does. That identity check runs below and fails loudly rather than silently
 * testing against the wrong function.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodeModule from 'node:module';
import * as NodePath from 'node:path';
import * as NodeUrl from 'node:url';

const [, , packageRootArgument] = process.argv;
if (!packageRootArgument) {
    process.stderr.write('usage: enumerate.mjs <tailwindcss package root>\n');
    process.exit(2);
}

const packageRoot = NodePath.resolve(packageRootArgument);
const packageJson = JSON.parse(NodeFileSystem.readFileSync(NodePath.join(packageRoot, 'package.json'), 'utf8'));

/*
 * Find the chunk holding the data-type predicates, by content rather than by filename. The chunk
 * hash changes with every Tailwind release, so hardcoding `chunk-C2OYBFIH.mjs` would break on the
 * next upgrade in a way that looks like a missing file rather than a stale assumption.
 */
const distributionDirectory = NodePath.join(packageRoot, 'dist');
const chunkCandidates = NodeFileSystem.readdirSync(distributionDirectory).filter((name) => name.endsWith('.mjs'));

let inferDataType = null;
let isLength = null;
let isPositiveInteger = null;
let segment = null;
let chunkName = null;

for (const candidate of chunkCandidates) {
    const candidateUrl = NodeUrl.pathToFileURL(NodePath.join(distributionDirectory, candidate)).href;
    let module;
    try {
        module = await import(candidateUrl);
    } catch {
        continue;
    }

    /*
     * Identify by behavior. A function is `inferDataType` when it takes (value, types), returns null
     * for a `var()` value regardless of the type list, returns 'color' for '#fff' when asked, and
     * returns null for '#fff' when not asked. That last clause is what separates it from a predicate
     * that happens to return a truthy string.
     */
    for (const exported of Object.values(module)) {
        if (typeof exported !== 'function' || exported.length !== 2) continue;
        try {
            if (exported('var(--a)', ['color', 'length']) !== null) continue;
            if (exported('#fff', ['color']) !== 'color') continue;
            if (exported('#fff', ['length']) !== null) continue;
            if (exported('10px', ['length']) !== 'length') continue;
            inferDataType = exported;
            chunkName = candidate;
        } catch {
            continue;
        }
    }

    if (inferDataType) {
        /*
         * `segment` takes (input, separator) like inferDataType takes (value, types), so arity does
         * not separate them. Identity is behavioral: it splits on a top-level separator, keeps
         * nested parens/brackets/braces together, and always pushes a final part, so an empty input
         * yields one empty string rather than none.
         */
        for (const exported of Object.values(module)) {
            if (typeof exported !== 'function' || exported.length !== 2 || exported === inferDataType) continue;
            try {
                const flat = exported('a,b,c', ',');
                if (!Array.isArray(flat) || flat.join('|') !== 'a|b|c') continue;
                if (exported('a(1,2),b', ',').join('|') !== 'a(1,2)|b') continue;
                if (exported('[a,b],c', ',').join('|') !== '[a,b]|c') continue;
                if (exported('', ',').length !== 1) continue;
                segment = exported;
            } catch {
                continue;
            }
        }

        for (const exported of Object.values(module)) {
            if (typeof exported !== 'function' || exported.length !== 1) continue;
            try {
                if (exported('10px') === true && exported('10') === false && exported('calc(1px)') === true) {
                    isLength = exported;
                }
                if (exported('12') === true && exported('1.5') === false && exported('+1') === false && exported('0') === true) {
                    isPositiveInteger = exported;
                }
            } catch {
                continue;
            }
        }
        break;
    }
}

if (!inferDataType) {
    process.stderr.write('could not locate inferDataType in any dist chunk\n');
    process.exit(1);
}

const allTypes = [
    'color', 'length', 'percentage', 'ratio', 'number', 'integer', 'url',
    'position', 'bg-size', 'line-width', 'image', 'family-name',
    'generic-name', 'absolute-size', 'relative-size', 'angle', 'vector',
];

const corpus = JSON.parse(NodeFileSystem.readFileSync(new URL('./corpus.json', import.meta.url), 'utf8'));

/*
 * Every value is asked three ways, because a single question cannot expose an ordering bug.
 *
 *   - Against the full list in declaration order. This is what a caller passing everything sees, and
 *     it is where first-match-wins shows up: `1 2 3` answers 'line-width', not 'vector'.
 *   - Against the full list reversed. A port that iterated its own fixed order instead of the
 *     caller's would agree on the first question and disagree here on every value matching more than
 *     one type. This is the question that catches it.
 *   - Against each single type alone. This isolates each predicate, so a wrong `isBackgroundSize`
 *     cannot hide behind an earlier type that also matched.
 */
const reversedTypes = [...allTypes].reverse();
const cases = [];

for (const value of corpus) {
    const perType = {};
    for (const type of allTypes) {
        perType[type] = inferDataType(value, [type]);
    }
    cases.push({
        value,
        all: inferDataType(value, allTypes),
        reversed: inferDataType(value, reversedTypes),
        perType,
        isLength: isLength ? isLength(value) : null,
        isPositiveInteger: isPositiveInteger ? isPositiveInteger(value) : null,
        segmentComma: segment ? segment(value, ',') : null,
        segmentSpace: segment ? segment(value, ' ') : null,
    });
}

process.stdout.write(JSON.stringify({
    tailwindVersion: packageJson.version,
    chunk: chunkName,
    isLengthResolved: Boolean(isLength),
    isPositiveIntegerResolved: Boolean(isPositiveInteger),
    segmentResolved: Boolean(segment),
    types: allTypes,
    reversedTypes,
    cases,
}, null, 2) + '\n');
