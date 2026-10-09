/*
 * Capture what the shipped Tailwind engine does with the repository's own `@utility` blocks, so the
 * Go evaluator in internal/tailwind/utility.go is tested against measurements rather than against a
 * reading of `createCssUtility`.
 *
 * The population is chosen to make the claim under test falsifiable rather than to be large:
 *
 *   - Every registry class whose root is defined by an `@utility` block. This is where the 18 known
 *     exceptions live, and it is the population Phase 0 measured, so agreement here is comparable
 *     against a number that already exists.
 *
 *   - A shape sweep over every such root: bare integers, theme keys of the `--percentage` and
 *     `--translate` namespaces, fractions, decimals, percentages, arbitrary values with and without
 *     a typehint, and modifier forms. The registry alone would not reach the ratio splice or the
 *     modifier post-conditions, because `getClassList()` advertises neither.
 *
 *   - The `@utility` blocks themselves, as parsed by the engine's own CSS parser and re-serialized,
 *     so the Go side evaluates the same tree rather than one it re-parsed from source with a
 *     parser that might differ.
 *
 * `compileAstNodes` is the ground truth, for the reason generate_descriptors states: it returns
 * `{order, count}` directly, which is the thing the Go side has to reproduce. A null reading is
 * recorded as null and never scored as agreement, because a candidate the engine rejects and one
 * that reads `[]#0` sort differently.
 *
 * Usage:
 *
 *   node internal/lint/rules/tailwind/tools/generate_utility/enumerate.mjs <theme.css>
 */

import NodeFileSystem from 'node:fs';
import NodePath from 'node:path';
import NodeUrl from 'node:url';
import NodeChildProcess from 'node:child_process';
import { loadDesignSystem, readingOf, parseCandidate } from '../generate_descriptors/loader.mjs';

const entryPointArgument = process.argv[2];
if (!entryPointArgument) {
    process.stderr.write('usage: enumerate.mjs <theme.css>\n');
    process.exit(2);
}

const { designSystem, tailwindVersion, entryPoint, recordedEntryPoint } = await loadDesignSystem(entryPointArgument);

/*
 * The theme namespaces, recovered the way generate_descriptor_table recovers them: every
 * candidate prefix of every key is offered to `keysInNamespaces`, which is the engine's own answer
 * about where a namespace boundary is rather than a guess from the string.
 */
const keysByNamespace = {};
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
        keysByNamespace[prefix] = keys;
    }
}

/*
 * The `@utility` blocks the repository declares, read from the stylesheet graph.
 *
 * Read from source rather than from the design system, because the design system does not expose
 * the blocks it registered: `designSystem.utilities` holds compiled closures, not the CSS they came
 * from. The import graph is walked the same way the loader walks it, and every `@utility` at-rule is
 * captured with its body serialized back to CSS.
 *
 * Serializing back to CSS rather than emitting the engine's AST is deliberate. The Go side has its
 * own parser (cssparser.go), already measured against this same engine, and having it parse the
 * block text is what keeps the two components composed the way they will be in production. Emitting
 * a pre-parsed tree would test the evaluator against a tree the Go parser might never produce.
 */
const utilityBlocks = [];
{
    const seen = new Set();
    const visit = (path) => {
        if (seen.has(path)) return;
        seen.add(path);
        let source;
        try {
            source = NodeFileSystem.readFileSync(path, 'utf8');
        }
        catch {
            return;
        }

        // Follow `@import`s so the whole graph is reached. Bare specifiers resolve through node,
        // relative ones against this file's directory.
        for (const match of source.matchAll(/@import\s+(?:url\()?['"]([^'"]+)['"]\)?[^;]*;/g)) {
            const specifier = match[1];
            let resolved = null;
            if (specifier.startsWith('.') || NodePath.isAbsolute(specifier)) {
                resolved = NodePath.resolve(NodePath.dirname(path), specifier);
            }
            else {
                continue;
            }
            if (NodeFileSystem.existsSync(resolved)) visit(resolved);
        }

        // `@utility <name> { ... }`, with brace matching so a nested rule does not end the block
        // early. The logical-direction utilities are exactly that shape.
        for (let index = 0; index < source.length; index++) {
            if (!source.startsWith('@utility', index)) continue;
            let cursor = index + '@utility'.length;
            while (cursor < source.length && /\s/.test(source[cursor])) cursor++;
            const nameStart = cursor;
            while (cursor < source.length && source[cursor] !== '{') cursor++;
            if (cursor >= source.length) break;
            const name = source.slice(nameStart, cursor).trim();
            const bodyStart = cursor;
            let depth = 0;
            for (; cursor < source.length; cursor++) {
                if (source[cursor] === '{') depth++;
                else if (source[cursor] === '}') { depth--; if (depth === 0) { cursor++; break; } }
            }
            utilityBlocks.push({ name, source: '@utility ' + name + ' ' + source.slice(bodyStart, cursor) });
            index = cursor - 1;
        }
    };
    visit(entryPoint);
}

const functionalBlocks = utilityBlocks.filter((block) => block.name.endsWith('-*'));
const functionalRoots = functionalBlocks.map((block) => block.name.slice(0, -2));
const functionalRootSet = new Set(functionalRoots);

/*
 * The classes to measure.
 *
 * Registry first, so the exception population is covered exactly as Phase 0 covered it, then a
 * shape sweep that reaches what the registry does not advertise: ratios, decimals, arbitrary
 * values, typehints, and modifiers.
 */
const classNames = new Set();
{
    for (const entry of designSystem.getClassList()) {
        const className = Array.isArray(entry) ? entry[0] : entry;
        const candidate = parseCandidate(designSystem, className);
        if (candidate?.kind === 'functional' && functionalRootSet.has(candidate.root)) classNames.add(className);
        // The bare root too, which is a separate static `@utility` in this repository and is the
        // control for post-condition 1: a functional block with no `--default(…)` must not compile
        // bare.
        if (candidate?.kind === 'static' && functionalRootSet.has(candidate.root)) classNames.add(className);
    }

    const percentageKeys = [];
    const translateKeys = [];
    for (const namespace of Object.keys(keysByNamespace)) {
        if (namespace.startsWith('--percentage')) percentageKeys.push(...keysByNamespace[namespace]);
        if (namespace === '--translate') translateKeys.push(...keysByNamespace[namespace]);
    }

    const shapes = [
        // Bare integers, on and off the `--percentage` scale.
        '0', '1', '2', '3', '4', '5', '7', '8', '11', '12', '16', '25', '50', '96', '100', '101',
        // Decimals: `1.25` is a valid spacing multiplier and `1.3` is not, which is the only thing
        // separating a surviving `--value(number)` from a dropped one.
        '0.5', '1.5', '2.5', '1.25', '1.3', '0.25', '0.75',
        // Fractions, which reach the ratio branch and the splice.
        '1/2', '2/3', '3/4', '9/12', '1/1', '0/1', '1/0', '1.5/2', '16/9',
        // Percentages, integer and not.
        '25%', '50%', '12.5%',
        // Arbitrary values, with and without a typehint.
        '[10px]', '[50%]', '[0.5]', '[1.5]', '[var(--a)]', '[length:var(--a)]', '[percentage:var(--a)]',
        '[color:var(--a)]', '[integer:4]', '[*]', '[calc(100%-1rem)]',
        // Keywords that are theme keys on some roots and nothing on others.
        'full', 'px', 'auto', 'none', 'translate-full',
    ];
    const modifierShapes = ['', '/50', '/none', '/[0.5]', '/[var(--a)]', '/2', '/3'];

    for (const root of functionalRoots) {
        for (const shape of shapes) {
            for (const modifier of modifierShapes) {
                classNames.add(root + '-' + shape + modifier);
            }
        }
        for (const key of [...percentageKeys, ...translateKeys]) {
            classNames.add(root + '-' + key);
            classNames.add(root + '-' + key + '/50');
        }
        classNames.add(root);
        classNames.add(root + '/50');
    }
}

/*
 * The measurement. `{order, count}` per class, plus the parsed candidate's value and modifier so a
 * Go-side disagreement can be read as "the candidate parsed differently" rather than only as "the
 * reading differed".
 */
const cases = [];
let nullReadings = 0;
for (const className of Array.from(classNames).sort()) {
    const candidate = parseCandidate(designSystem, className);
    const reading = readingOf(designSystem, className);
    if (reading === null) nullReadings++;
    cases.push({
        className,
        candidateKind: candidate?.kind ?? null,
        root: candidate?.root ?? null,
        valueKind: candidate?.value?.kind ?? null,
        value: candidate?.value?.value ?? null,
        valueDataType: candidate?.value?.dataType ?? null,
        fraction: candidate?.value?.fraction ?? null,
        modifierKind: candidate?.modifier?.kind ?? null,
        modifier: candidate?.modifier?.value ?? null,
        reading: reading === null ? null : { order: reading.order, count: reading.count },
    });
}

/*
 * The per-declaration roots, measured the way generate_descriptor_table measures them: probe a
 * root with a value that is only an integer and with one that is also a theme key, and a root whose
 * count changes between them is resolving per declaration. Carried so the Go side can assert it
 * finds the same set, rather than being told which roots to treat specially.
 */
const perDeclarationRoots = [];
{
    const percentageKeys = new Set();
    for (const namespace of Object.keys(keysByNamespace)) {
        if (!namespace.startsWith('--percentage')) continue;
        for (const key of keysByNamespace[namespace]) percentageKeys.add(key);
    }
    for (const root of [...functionalRoots].sort()) {
        const integerOnly = readingOf(designSystem, root + '-4');
        if (integerOnly === null) continue;
        for (const key of percentageKeys) {
            const both = readingOf(designSystem, root + '-' + key);
            if (both === null) continue;
            if (both.count !== integerOnly.count) { perDeclarationRoots.push(root); break; }
        }
    }
}

/*
 * The 18: the registry classes the descriptor model mispredicts.
 *
 * Not re-derived here. `extract.mjs` is the tool that measured the model against the engine over the
 * whole registry, and its `disagreeingRoots` is the definition of the exception set; a second
 * definition written here would be a proxy for it, and a proxy that agrees on this repository is
 * exactly the kind of right-answer-for-the-wrong-reason this component is supposed to refuse.
 *
 * A first attempt did write one, "a registry class whose count differs from what a single-path value
 * produces", and it reported 30 rather than 18: `<root>-full` also resolves through two paths, and
 * the descriptor model answers it correctly through the `--translate` namespace. The proxy was
 * measuring the mechanism, while the exception set is about which readings the *table* gets wrong,
 * and only the table's own prediction can say that.
 *
 * So the extractor is run and its report is read. It takes a few minutes, and a manual generator is
 * the right place to pay that.
 */
const knownExceptions = [];
const shadowQuirkProbes = [];
{
    const extractPath = NodePath.join(NodePath.dirname(NodeUrl.fileURLToPath(import.meta.url)), '..', 'generate_descriptors', 'extract.mjs');
    const report = JSON.parse(NodeChildProcess.execFileSync(process.execPath, [extractPath, entryPoint], {
        encoding: 'utf8',
        maxBuffer: 1024 * 1024 * 256,
    }));
    for (const root of report.disagreeingRoots ?? []) {
        for (const example of root.examples ?? []) {
            // The sweep half of the report carries `group`; the registry half carries `kind`. Only
            // the registry classes are the 18. The four sweep probes are the other exception, the
            // upstream shadow quirk on `[16/9]`, captured separately below.
            if (!example.kind) continue;
            const reading = readingOf(designSystem, example.className);
            if (reading === null) continue;
            knownExceptions.push({
                className: example.className,
                order: reading.order,
                count: reading.count,
                predictedByDescriptorModel: example.predicted,
            });
        }
    }
    knownExceptions.sort((left, right) => (left.className < right.className ? -1 : 1));

    /*
     * The other exception, recorded rather than modelled: four sweep probes hitting an upstream
     * shadow quirk on `[16/9]` values, where the shadow handler splits an arbitrary value on `/`
     * looking for a colour slot and emits garbage CSS. No registry contains them and no `@utility`
     * block is involved, so they are outside this component; carried so the boundary is stated in
     * the fixture rather than only in prose.
     */
    for (const root of report.disagreeingRoots ?? []) {
        for (const example of root.examples ?? []) {
            if (example.kind) continue;
            shadowQuirkProbes.push({
                className: example.className,
                predicted: example.predicted,
                actual: example.actual,
            });
        }
    }
    shadowQuirkProbes.sort((left, right) => (left.className < right.className ? -1 : 1));
}

/*
 * The synthetic half: `@utility` blocks reaching the branches this repository's own blocks never do.
 *
 * The repository half is the exit criterion and it is not sufficient on its own. Its blocks use four
 * argument forms, `integer`, `number`, `ratio` and theme namespaces, plus bracketed types. Four
 * branches of `resolveValueFunction` are therefore unreachable from it, and each was found by
 * mutating the Go and watching nothing fail:
 *
 *   - `--value(percentage)` and its integer-percentage guard. Deleting the guard changed no reading
 *     in the repository corpus, because no block names `percentage` as a bare type at all.
 *   - `--modifier(…)`, which the four post-conditions gate. No repository block uses one, so every
 *     modifier class in the repository corpus is rejected by post-condition 4 without the modifier
 *     resolver ever running.
 *   - `--default(…)`, the omitted-value path. Without it a functional block does not compile bare,
 *     which is why this repository writes a separate static `@utility` for each bare form, and why
 *     the branch that would make that unnecessary is never taken.
 *   - Quoted literals, `--value('closest-side')`.
 *
 * Each is a real code path in the port, measured here against the engine rather than left as surface
 * that reads as supported. A branch with no test is worse than an absent branch, because it looks
 * like coverage in a diff.
 *
 * Every synthetic design system is loaded standalone, with its own tiny `@theme`, so the readings are
 * a function of the stylesheet in front of it and nothing else.
 */
const syntheticCases = [];
const syntheticStylesheets = [
    [
        'percentage-and-integer',
        `@theme { --spacing: 0.25rem; --percentage-half: 0.5; }
         @utility pct-* {
             --a: calc(--value(percentage) * 1);
             --a: calc(--value(integer) * var(--spacing));
             --a: --value(--percentage-*);
         }`,
        ['pct-50%', 'pct-12.5%', 'pct-25%', 'pct-0%', 'pct-100%', 'pct-4', 'pct-half', 'pct-1.5', 'pct-[10px]', 'pct-50%/2', 'pct'],
    ],
    [
        'modifier',
        `@theme { --spacing: 0.25rem; --leading-none: 1; }
         @utility mod-* {
             --a: --value(integer);
             --b: --modifier(integer);
         }`,
        ['mod-4', 'mod-4/2', 'mod-4/none', 'mod-4/2.5', 'mod-4/[3]', 'mod-2/3', 'mod', 'mod-abc', 'mod-abc/2'],
    ],
    [
        'modifier-with-theme',
        `@theme { --leading-none: 1; --percentage-half: 0.5; }
         @utility modt-* {
             --a: --value(--percentage-*);
             --b: --modifier(--leading-*);
         }`,
        ['modt-half', 'modt-half/none', 'modt-half/50', 'modt-half/[0.5]', 'modt-none'],
    ],
    [
        'default',
        `@theme { --spacing: 0.25rem; }
         @utility def-* {
             --a: --value(integer, --default(7));
         }`,
        ['def', 'def-4', 'def-abc', 'def-4/2', 'def/2'],
    ],
    [
        'literal',
        `@theme { --spacing: 0.25rem; }
         @utility lit-* {
             --a: --value('closest-side');
             --a: --value('farthest-side');
             --a: --value(integer);
         }`,
        ['lit-closest-side', 'lit-farthest-side', 'lit-4', 'lit-other', 'lit-closest-side/2'],
    ],
    [
        'ratio-and-modifier',
        `@theme { --spacing: 0.25rem; }
         @utility rat-* {
             --a: calc(--value(ratio));
             --b: calc(--value(integer) * var(--spacing));
             --c: --modifier(integer);
         }`,
        ['rat-1/2', 'rat-4', 'rat-4/2', 'rat-2/3', 'rat-1.5/2', 'rat-0/1', 'rat'],
    ],
    [
        // The unnormalized argument spellings. Every `--value()` argument in both repositories is
        // authored already normalized, so normalization changes nothing there but comma spacing and
        // deleting the whole pass changes no reading. These spellings are what it is actually for:
        // `--value(--percentage)` names a namespace only after the `-*` is appended, and
        // `--value(--text --line-height)` is the nested form written with a space.
        'unnormalized-arguments',
        `@theme { --percentage-half: 0.5; --text-sm: 0.875rem; --text-sm--line-height: 1.25rem; }
         @utility unnorm-* {
             --a: --value(--percentage);
             --b: --value(--text --line-height);
             --c: --value([ *]);
         }`,
        ['unnorm-half', 'unnorm-sm', 'unnorm-[10px]', 'unnorm-4', 'unnorm-nope'],
    ],
    [
        // A `--value()` naming a data type outside `BARE_VALUE_DATA_TYPES`. Upstream skips the
        // argument entirely and warns, which is what stops `@utility x-* { color: --value(color) }`
        // from making `x-#0088cc` a valid class. Without the allowlist a port would infer the type
        // and resolve, so the class would compile where the engine rejects it.
        //
        // `length` is used rather than `color` because the point is the allowlist and not the colour
        // path: `4px` is a length, so a port without the guard resolves it and this case separates
        // the two. The `integer` declaration is there so the utility is otherwise valid, which is
        // what makes the difference a count rather than a rejection.
        'disallowed-bare-type',
        `@theme { --spacing: 0.25rem; }
         @utility bad-* {
             --a: --value(length);
             --b: --value(integer);
         }`,
        ['bad-4px', 'bad-4', 'bad-2rem', 'bad-abc'],
    ],
    [
        // A block using `--modifier(…)` and no `--value(…)` at all. Upstream's post-condition 1 is
        // `if (!usedValueFn || !resolvedValueFn) return null`, and `usedValueFn` is set only by
        // `--value(…)`, so this utility never compiles no matter what the modifier resolves to.
        //
        // It exists because the two flags are easy to conflate: a port that let `--modifier(…)` set
        // `usedValueFunction` would compile every one of these, and no block in either repository
        // uses a modifier at all, so nothing else separates the two spellings.
        'modifier-without-value',
        `@theme { --spacing: 0.25rem; }
         @utility modonly-* {
             --a: --modifier(integer);
         }`,
        ['modonly-4/2', 'modonly-4', 'modonly/2', 'modonly'],
        // Rejection-only: post-condition 1 rejects every probe, which is the whole point of the
        // case, so the usual "some probe must compile" assertion does not apply to it.
        { rejectionOnly: true },
    ],
    [
        'nested-theme-key',
        `@theme { --text-sm: 0.875rem; --text-sm--line-height: 1.25rem; }
         @utility nest-* {
             --a: --value(--text-*--line-height);
             --b: --value(--text-*);
         }`,
        ['nest-sm', 'nest-lg', 'nest-4'],
    ],
    [
        'arbitrary-typehints',
        `@theme { --spacing: 0.25rem; }
         @utility arb-* {
             --a: --value([length]);
             --b: --value([percentage]);
         }`,
        ['arb-[10px]', 'arb-[50%]', 'arb-[length:var(--x)]', 'arb-[color:var(--x)]', 'arb-[var(--x)]', 'arb-4'],
    ],
    [
        'any-arbitrary',
        `@theme { --spacing: 0.25rem; }
         @utility any-* {
             --a: --value([*]);
         }`,
        ['any-[10px]', 'any-[anything]', 'any-[color:var(--x)]', 'any-4', 'any'],
    ],
];

for (const [name, stylesheet, probes, options = {}] of syntheticStylesheets) {
    // Written beside the repository's own entry point rather than in a temporary directory, so that
    // `@import "tailwindcss"` resolves against the same install the repository half measured. A
    // stylesheet in the system temp directory has no `node_modules` above it and the import fails.
    const temporaryPath = NodePath.join(NodePath.dirname(entryPoint), `.gen-tailwind-utility-${name}.css`);
    NodeFileSystem.writeFileSync(temporaryPath, '@import "tailwindcss";\n' + stylesheet);

    let syntheticSystem;
    try {
        ({ designSystem: syntheticSystem } = await loadDesignSystem(temporaryPath));
    }
    finally {
        NodeFileSystem.rmSync(temporaryPath, { force: true });
    }
    const blocks = [];
    for (const match of stylesheet.matchAll(/@utility\s+([^\s{]+)\s*\{/g)) {
        const start = match.index;
        let depth = 0;
        let cursor = start;
        for (; cursor < stylesheet.length; cursor++) {
            if (stylesheet[cursor] === '{') depth++;
            else if (stylesheet[cursor] === '}') { depth--; if (depth === 0) { cursor++; break; } }
        }
        blocks.push({ name: match[1], source: stylesheet.slice(start, cursor) });
    }

    const themeEntries = {};
    for (const [key, value] of syntheticSystem.theme.entries()) themeEntries[key] = value.value ?? value;

    const cases = [];
    for (const probe of probes) {
        let candidate = null;
        try { candidate = Array.from(syntheticSystem.parseCandidate(probe))[0] ?? null; } catch { candidate = null; }
        let reading = null;
        if (candidate) {
            try {
                const nodes = syntheticSystem.compileAstNodes(candidate);
                const propertySort = nodes?.[0]?.propertySort;
                if (propertySort) reading = { order: propertySort.order ?? [], count: propertySort.count ?? 0 };
            }
            catch { reading = null; }
        }
        cases.push({
            className: probe,
            candidateKind: candidate?.kind ?? null,
            root: candidate?.root ?? null,
            valueKind: candidate?.value?.kind ?? null,
            value: candidate?.value?.value ?? null,
            valueDataType: candidate?.value?.dataType ?? null,
            fraction: candidate?.value?.fraction ?? null,
            modifierKind: candidate?.modifier?.kind ?? null,
            modifier: candidate?.modifier?.value ?? null,
            reading,
        });
    }

    syntheticCases.push({ name, stylesheet, rejectionOnly: options.rejectionOnly === true, themeEntries, utilityBlocks: blocks, cases });
}

process.stdout.write(JSON.stringify({
    tailwindVersion,
    entryPoint: recordedEntryPoint,
    keysByNamespace,
    utilityBlocks,
    perDeclarationRoots,
    knownExceptions,
    shadowQuirkProbes,
    syntheticCases,
    nullReadings,
    cases,
}, null, 2));
