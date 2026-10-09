/*
 * Capture the compiled declarations of every multi-declaration framework root, per class.
 *
 * Waves 1 and 2a were tables of `{property}` because each root emitted exactly one declaration and
 * the property was readable from the source. This wave is not that, and an attempt to make it one is
 * what produced this file.
 *
 * # Why the source shape was not enough
 *
 * Twenty-one of these roots match a single pattern in `utilities.ts`: a properties wrapper, a
 * `--tw-*` declaration, and a final declaration of the real property. Extracting that pattern
 * suggested they all read the same way. They do not. Measured against the engine:
 *
 *     grayscale     [334,339]#2      one --tw- var plus filter
 *     skew-3        [69,70,71]#3     writes both --tw-skew-x and --tw-skew-y
 *     skew-x-3      [69,71]#2        writes one
 *     ease-linear   [354]#2          two declarations sharing one property position
 *
 * The wrapper is `atRoot([...])`, which PropertySort does not descend into, so it contributes
 * nothing to a reading. That part the source does predict. What it does not predict is how many
 * declarations survive underneath, and `ease` is the case that proves it: two declarations at one
 * order index, so a table keyed on the property name would have said `#1`.
 *
 * So this asks the engine for the declarations rather than inferring them from the handler's text.
 * A root's reading is captured per class, because arity here is not a property of the root: a
 * default value, a theme hit and a bare value can each produce a different count.
 *
 * Usage:
 *
 *   node internal/lint/rules/tailwind/tools/generate_functional/wave2c.mjs <theme.css> [--resolve-root <dir>] > wave2b.json
 */

import { loadDesignSystem, parseCandidate, readingOf } from '../generate_descriptors/loader.mjs';

const entryPointArgument = process.argv[2];
if (!entryPointArgument) {
    process.stderr.write('usage: wave2c.mjs <theme.css> [--resolve-root <dir>]\n');
    process.exit(2);
}

function flagValue(name) {
    const index = process.argv.indexOf(name);
    return index === -1 ? undefined : process.argv[index + 1];
}

const { designSystem, tailwindVersion, entryPoint, recordedEntryPoint } = await loadDesignSystem(
    entryPointArgument,
    flagValue('--resolve-root'),
);

const roots = [
    'aspect', 'auto-cols', 'auto-rows', 'col-span', 'content', 'divide-x', 'divide-y',
    'font-stretch', 'grid-cols', 'grid-rows', 'line-clamp', 'mask-conic', 'mask-linear',
    'mask-radial', 'row-span', 'tracking', 'transition', 'underline-offset',
];

/*
 * The probe suffixes, and why the fraction ones are here.
 *
 * `-16/9` and `-3/2` were added after a mutation survived. `aspect` reads its bare value off
 * `Fraction` rather than `Value`, and no probe in the original set carried a slash, so removing the
 * fraction guard entirely changed no answer anywhere. The engine reads `aspect-16/9` as `[41]#1`,
 * and adding the probes turned a table row that silently declined into a measured one.
 */
const probeSuffixes = ['', '-4', '-0', '-1', '-50', '-90', '-100', '-sm', '-none', '-linear',
    '-[3px]', '-[var(--a)]', '-2.5', '-1.3', '/50', '-16/9', '-3/2', '-1.3/2', '-7/3'];

const registryByRoot = new Map();
for (const entry of designSystem.getClassList?.() ?? []) {
    const name = Array.isArray(entry) ? entry[0] : typeof entry === 'string' ? entry : entry?.name;
    if (typeof name !== 'string') continue;
    const candidate = parseCandidate(designSystem, name);
    if (!candidate || candidate.kind !== 'functional' || !candidate.root) continue;
    if (!roots.includes(candidate.root)) continue;
    const list = registryByRoot.get(candidate.root) ?? [];
    list.push(name);
    registryByRoot.set(candidate.root, list);
}

const cases = [];
for (const root of roots) {
    const names = new Set(registryByRoot.get(root) ?? []);
    for (const suffix of probeSuffixes) names.add(root + suffix);
    names.add('-' + root + '-3');

    for (const className of names) {
        const candidate = parseCandidate(designSystem, className);
        const reading = readingOf(designSystem, className);
        cases.push({
            root,
            className,
            parsedRoot: candidate?.root ?? null,
            reading: reading === null ? null : { order: reading.order, count: reading.count },
        });
    }
}

const answered = cases.filter((entry) => entry.reading !== null).length;
const rejected = cases.length - answered;
if (answered === 0 || rejected === 0) {
    process.stderr.write('wave2c.mjs: one side of the population is empty, so a comparison over this fixture would prove nothing\n');
    process.exit(4);
}

/*
 * The distinct readings per root, which is the number this wave turns on.
 *
 * A root with one distinct reading is a table row. A root with several is a root whose arity depends
 * on its value, and it needs its own representation rather than a row.
 */
const distinctByRoot = {};
for (const root of roots) {
    const seen = new Set();
    for (const entry of cases) {
        if (entry.root !== root || entry.reading === null) continue;
        seen.add(JSON.stringify(entry.reading));
    }
    distinctByRoot[root] = [...seen].map((raw) => JSON.parse(raw));
}

process.stdout.write(JSON.stringify({
    tailwindVersion,
    entryPoint: recordedEntryPoint,
    groundTruth: 'compileAstNodes propertySort, via readingOf',
    rootCount: roots.length,
    caseCount: cases.length,
    answered,
    rejected,
    distinctByRoot,
    cases,
}, null, 2) + '\n');
