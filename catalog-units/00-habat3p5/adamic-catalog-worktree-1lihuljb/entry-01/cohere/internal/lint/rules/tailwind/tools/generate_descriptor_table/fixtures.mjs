/*
 * The engine's answers, for the Go lookup to be measured against.
 *
 * The Go test needs three things per case and nothing else: the class, the parse the lookup would
 * receive, and the `{order, count}` the engine produced. It deliberately does not store parse trees
 * or intermediate values. That restraint is the point rather than a preference: an earlier fixture
 * in this slice reached 6.5 MB by storing trees it did not need, and the trees made it no more able
 * to fail.
 *
 * The parse is included because `internal/tailwind`'s candidate parser (#bg0t9tb) is a separate task
 * and the descriptor lookup must be testable without it. Carrying the engine's own parse means a
 * lookup failure is a lookup failure rather than a parser failure wearing its clothes, which is what
 * lets this suite report on the model rather than on the sum of two ports.
 *
 * Usage:
 *
 *   node internal/lint/rules/tailwind/tools/generate_descriptor_table/fixtures.mjs <theme.css> [--limit N] > fixtures.json
 */

import { loadDesignSystem, parseCandidate, readingOf } from '../generate_descriptors/loader.mjs';

const [entryPointArgument, ...restArguments] = process.argv.slice(2);
if (!entryPointArgument) {
    process.stderr.write('usage: fixtures.mjs <theme.css> [--limit N]\n');
    process.exit(2);
}
const limitFlagIndex = restArguments.indexOf('--limit');
const limit = limitFlagIndex >= 0 ? Number(restArguments[limitFlagIndex + 1]) : Infinity;

const { designSystem, tailwindVersion, entryPoint, recordedEntryPoint } = await loadDesignSystem(entryPointArgument);

/*
 * The populations, and why there is more than one.
 *
 * The registry is what the engine says it supports and is the headline number. It is also, by
 * construction, made only of classes that resolve, so on its own it never exercises the paths that
 * decline. The arbitrary shapes are what put values through inference rather than through the theme,
 * and they are where the per-root type ordering is observable at all: `bg-[3px]` is the canary for
 * `position` before `length`, and no registry class can show it.
 */
const classNames = new Set();
const modifierClassNames = [];
for (const entry of designSystem.getClassList()) {
    const name = Array.isArray(entry) ? entry[0] : entry;
    classNames.add(name);
    if (Array.isArray(entry) && entry[1] && typeof entry[1] === 'object') {
        for (const modifier of entry[1].modifiers ?? []) modifierClassNames.push(name + '/' + modifier);
    }
}

// Modifiers are the third axis and the registry keeps them separate, so they are sampled in rather
// than folded into the headline count. Every eleventh, which is a stride rather than a random sample
// so the corpus is the same on every run and a diff means a real change.
for (let index = 0; index < modifierClassNames.length; index += 11) {
    classNames.add(modifierClassNames[index]);
}

/*
 * The arbitrary shapes, applied to every functional root.
 *
 * This is the population that separates the orderings. `[3px]` is a length and a position; `[red]`
 * is a colour and a family-name; `[16/9]` is a ratio and the value that trips the upstream shadow
 * quirk. The annotated forms are here because the engine consumes the annotation and infers on what
 * remains, which is a split the Go side has to reproduce and which no bare value can test.
 */
const arbitraryShapes = [
    '[3px]', '[red]', '[16/9]', '[50%]', '[2]', '[0.5]', '[var(--a)]', '[url(a.png)]',
    '[color:var(--a)]', '[length:3px]', '[url:x]', '[position:top]', '[image:none]',
    '[1fr]', '[calc(1px+2px)]', '[#fff]', '[oklch(0.5_0.2_30)]', '[serif]', '[45deg]',
    '[3px]/50', '[3px]/none', '[3px]/[0.5]', '[red]/50', '[red]/none', '[var(--a)]/[var(--b)]',
];
const bareShapes = ['4', 'full', 'none', 'bold', 'current', 'inherit', 'transparent', '25%', '1/2', 'sm', 'auto', '0'];

const functionalRoots = new Set();
for (const className of classNames) {
    for (const probe of [className, className + '-4']) {
        const candidate = parseCandidate(designSystem, probe);
        if (candidate?.kind === 'functional' && candidate.root) functionalRoots.add(candidate.root);
    }
}
for (const root of functionalRoots) {
    for (const shape of arbitraryShapes) classNames.add(root + '-' + shape);
    for (const shape of bareShapes) classNames.add(root + '-' + shape);
    // The empty form, which some roots accept and others reject, and the two are different facts.
    classNames.add(root);
    classNames.add(root + '/50');
}

// A handful of arbitrary properties, which are not utilities at all and take the other branch.
for (const property of ['[font:inherit]', '[color:red]', '[--custom:1]', '[display:flex]', '[nonsense:1]']) {
    classNames.add(property);
}

/*
 * Reduction by lookup path, because 104,220 cases hold 5,353 distinct ones.
 *
 * The full population is generated and then reduced rather than sampled, so the reduction is a
 * measurement of redundancy rather than a guess about which cases matter. The first run of this tool
 * wrote 39 MB: the registry is 37,643 classes and most of them walk the same path, since every
 * `--color` key on `bg` proves exactly what the first one proved.
 *
 * The signature is what the lookup actually branches on: the root, the value's kind and annotation,
 * the modifier axis, and the reading produced. Two cases with the same signature cannot disagree in
 * the Go port unless they also disagree in the engine, and if they do, the reading is in the
 * signature and they are kept.
 *
 * Bare values need more than the signature, and less than all of them.
 *
 * Their path runs through the theme-namespace scan, so which *key* was written decides which
 * namespace claimed it, and that is invisible to a reading-based signature: `font-bold` and
 * `font-serif` share one and are exactly the pair that separates the two precedence orderings that
 * refinement 2 exists to pin. But keeping every named value keeps 37,643 registry classes to prove
 * something the first few keys of each namespace already prove, which is how the first run of this
 * tool wrote 39 MB.
 *
 * So a bare case is kept per (root, claiming namespace), plus a fixed number of distinct keys from
 * each, rather than per key. The claiming namespace is computed here with the engine's own key sets
 * in the engine's own longest-first order, which is the same scan the Go lookup runs. That makes the
 * kept population "every way a bare value can resolve on every root" instead of "every value", and
 * it is the difference between a corpus that measures the model and one that measures the theme.
 */
const namespacesLongestFirst = [];
const keysByNamespaceSet = new Map();
{
    const candidatePrefixes = new Set();
    for (const [key] of designSystem.theme.entries()) {
        const segments = key.slice(2).split('-');
        for (let length = 1; length <= segments.length - 1; length++) {
            candidatePrefixes.add('--' + segments.slice(0, length).join('-'));
        }
    }
    for (const prefix of candidatePrefixes) {
        let keys;
        try {
            keys = designSystem.theme.keysInNamespaces([prefix]);
        }
        catch {
            continue;
        }
        if (!keys || keys.length === 0) continue;
        keysByNamespaceSet.set(prefix, new Set(keys));
    }
    namespacesLongestFirst.push(
        ...Array.from(keysByNamespaceSet.keys()).sort((left, right) => right.length - left.length || (left < right ? -1 : 1)),
    );
}

function claimingNamespace(value) {
    for (const namespace of namespacesLongestFirst) {
        if (keysByNamespaceSet.get(namespace).has(value)) return namespace;
    }
    return '@none';
}

// Enough keys per (root, namespace) that a rule keyed on one key rather than on the namespace shows
// up, and few enough that the corpus stays a corpus. Four is arbitrary in the way a sample size is
// arbitrary; what is not arbitrary is that it is per namespace rather than per root.
const bareKeysPerBucket = 4;
const bareBucketCounts = new Map();

const keptSignatures = new Set();
function shouldKeep(entry) {
    if (entry.valueKind === 'named') {
        const bucket = [
            entry.root,
            claimingNamespace(entry.value),
            entry.modifierKind === 'none' ? 'absent' : entry.modifierValue === 'none' ? 'themed' : 'alpha',
            entry.reading === null ? 'null' : entry.reading.order.join('.') + '#' + entry.reading.count,
        ].join('|');
        const seen = bareBucketCounts.get(bucket) ?? 0;
        if (seen >= bareKeysPerBucket) return false;
        bareBucketCounts.set(bucket, seen + 1);
        return true;
    }
    const signature = [
        entry.kind, entry.root, entry.valueKind, entry.dataType,
        entry.modifierKind === 'none' ? 'absent' : entry.modifierValue === 'none' ? 'themed' : 'alpha',
        entry.reading === null ? 'null' : entry.reading.order.join('.') + '#' + entry.reading.count,
    ].join('|');
    if (keptSignatures.has(signature)) return false;
    keptSignatures.add(signature);
    return true;
}

const cases = [];
let nullCount = 0;
for (const className of Array.from(classNames).sort()) {
    if (cases.length >= limit) break;
    const candidate = parseCandidate(designSystem, className);
    const reading = readingOf(designSystem, className);
    /*
     * A class with no reading is recorded, never scored.
     *
     * The engine returns nothing for classes it cannot compile, and a harness that counted those as
     * agreement would be blind on exactly the population most likely to break. They are kept in the
     * fixture with a null reading so the Go side can assert it also declines, which is a real
     * assertion rather than a skipped one.
     */
    if (reading === null) nullCount++;
    if (!candidate) continue;

    const entry = {
        className,
        kind: candidate.kind,
        root: candidate.root ?? '',
        property: candidate.property ?? '',
        valueKind: candidate.value ? candidate.value.kind : 'none',
        value: candidate.value ? String(candidate.value.value ?? '') : '',
        /*
         * The fraction, which the engine reports and this fixture used to drop.
         *
         * `-bottom-1/2` parses with value `1`, modifier `2` AND fraction `1/2`, and the fraction is
         * what makes it resolve: without it the modifier cancels the bare value and the class reads
         * nothing. Recording only the first two produced a candidate no parser would ever build, so a
         * consumer that resolves a value declined classes the engine reads while a consumer that
         * carries a pre-measured reading did not.
         *
         * That difference blocked the base table's deletion: the two paths were not comparable on
         * this corpus, and the gap was in the fixture rather than in either path.
         */
        fraction: candidate.value?.fraction ?? '',
        dataType: candidate.value?.dataType ?? '',
        modifierKind: candidate.modifier ? candidate.modifier.kind : 'none',
        modifierValue: candidate.modifier ? String(candidate.modifier.value ?? '') : '',
        reading: reading === null ? null : { order: reading.order, count: reading.count },
    };
    if (shouldKeep(entry)) cases.push(entry);
}

process.stdout.write(JSON.stringify({
    tailwindVersion,
    entryPoint: recordedEntryPoint,
    counts: {
        cases: cases.length,
        classesConsidered: classNames.size,
        roots: functionalRoots.size,
        withoutReading: cases.filter((one) => one.reading === null).length,
        nullReadingsSeen: nullCount,
    },
    cases,
}, null, 2) + '\n');
