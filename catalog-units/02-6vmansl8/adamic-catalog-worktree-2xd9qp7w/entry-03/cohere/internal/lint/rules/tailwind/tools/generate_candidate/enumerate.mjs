/*
 * Ask the real Tailwind 4.3.3 candidate parser what it read for a corpus of class strings, and emit
 * the candidates as JSON.
 *
 * The Go port of `candidate.ts` is checked against these readings rather than against a reading of
 * the TypeScript source, for the reason generate_css_parser and generate_data_type exist: the
 * two are different claims. The one that decides this component is **order**. `parseCandidate` is a
 * generator, several classes yield more than one candidate, and the first that compiles wins. A
 * port that yields the same set in a different order gives a different final answer with nothing
 * anywhere reporting an error. `border-b` is the canonical case: `border-b` as an exact functional
 * root comes first, `border` with the named value `b` comes second. Swap them and every
 * `border-b` in the repository silently resolves to a different utility.
 *
 * So each case records the candidates as an ordered array, and `ambiguous` marks the ones where
 * that order is observable at all.
 *
 * # Why the design-system tables are captured too
 *
 * `parseCandidate` is not a pure function of its input. It asks the design system four questions:
 * `utilities.has(root, kind)`, `variants.has(root)`, `variants.kind(root)` and
 * `variants.compoundsWith(root, child)`. Those answers are what decide where a class splits, and
 * they come from this repository's `@theme` and `@utility` blocks as much as from the framework.
 *
 * Capturing the answers alongside the readings is what makes the Go test a test of the parser
 * rather than a test of a stubbed table someone hand-typed. The port is handed the same 1,201
 * utility roots and 88 variant roots the engine had, so a disagreement is a parser defect and can
 * be nothing else.
 *
 * # Why the public API, not a bundle probe
 *
 * generate_css_parser has to re-export the bundle's source and locate `parse` by behavior,
 * because `parse` is not exported. `parseCandidate` is different: `__unstable__loadDesignSystem` is
 * a real export and it returns an object carrying `parseCandidate` under that name, surviving
 * minification because it is a property rather than a binding. The identity check below still runs
 * — `border-b` must yield exactly the two-candidate reading described above — so a future Tailwind
 * that reshapes this fails loudly instead of quietly capturing something else.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodePath from 'node:path';

/* Empty positional slots are how the Go side passes "not supplied" for an earlier argument. */
const [, , packageRootArgument, corpusRootRaw, themeEntryRaw] = process.argv;
const corpusRootArgument = corpusRootRaw || null;
const themeEntryArgument = themeEntryRaw || null;
if (!packageRootArgument) {
    process.stderr.write('usage: enumerate.mjs <tailwindcss package root> [corpus root] [theme entry css]\n');
    process.exit(2);
}

const packageRoot = NodePath.resolve(packageRootArgument);
const packageJson = JSON.parse(NodeFileSystem.readFileSync(NodePath.join(packageRoot, 'package.json'), 'utf8'));

const { __unstable__loadDesignSystem } = await import(
    NodePath.join(packageRoot, 'dist', 'lib.mjs')
);

/*
 * Load the design system. When a repository theme entry point is named, that file is the entry and
 * its `@import` graph is followed, so the tables carry this repository's own `@utility` blocks. The
 * exit criterion asks for the real repository, not for framework defaults.
 */
const themeEntry = themeEntryArgument ? NodePath.resolve(themeEntryArgument) : null;
const entrySource = themeEntry
    ? NodeFileSystem.readFileSync(themeEntry, 'utf8')
    : '@import "tailwindcss";';
const entryBase = themeEntry ? NodePath.dirname(themeEntry) : packageRoot;

async function loadStylesheet(id, base) {
    if (id === 'tailwindcss') {
        return { base: packageRoot, content: NodeFileSystem.readFileSync(NodePath.join(packageRoot, 'index.css'), 'utf8') };
    }
    if (id.startsWith('tailwindcss/')) {
        const relative = id.slice('tailwindcss/'.length);
        for (const candidate of [relative, `${relative}.css`]) {
            const full = NodePath.join(packageRoot, candidate);
            if (NodeFileSystem.existsSync(full)) {
                return { base: NodePath.dirname(full), content: NodeFileSystem.readFileSync(full, 'utf8') };
            }
        }
    }
    const resolved = NodePath.resolve(base, id);
    for (const candidate of [resolved, `${resolved}.css`]) {
        if (NodeFileSystem.existsSync(candidate)) {
            return { base: NodePath.dirname(candidate), content: NodeFileSystem.readFileSync(candidate, 'utf8') };
        }
    }
    // A stylesheet the loader cannot find contributes nothing rather than aborting the capture; the
    // volume assertions at the end are what catch a graph that mostly failed to load.
    return { base, content: '' };
}

const designSystem = await __unstable__loadDesignSystem(entrySource, {
    base: entryBase,
    loadStylesheet,
    loadModule: async () => ({ base: entryBase, module: {} }),
});

/*
 * Identity check. `parseCandidate` must be the ambiguity-producing generator described at the top,
 * not something that merely returns an array of one.
 */
{
    const probe = [...designSystem.parseCandidate('border-b')];
    const shapedRight =
        probe.length === 2 &&
        probe[0].kind === 'functional' && probe[0].root === 'border-b' && probe[0].value === null &&
        probe[1].kind === 'functional' && probe[1].root === 'border' &&
        probe[1].value?.kind === 'named' && probe[1].value.value === 'b';
    if (!shapedRight) {
        process.stderr.write(
            'parseCandidate did not produce the known two-candidate reading for `border-b`; ' +
                'the export was located but is not the parser this fixture describes\n',
        );
        process.exit(3);
    }
}

/*
 * The design-system tables the parser consults, serialized so the Go test can answer the same four
 * questions the engine answered without reimplementing the theme.
 */
const utilityRoots = {};
for (const [root, entries] of designSystem.utilities.utilities) {
    utilityRoots[root] = [...new Set(entries.map((entry) => entry.kind))].sort();
}

const variantRoots = {};
for (const [root, info] of designSystem.variants.variants) {
    variantRoots[root] = { kind: info.kind, compounds: info.compounds, compoundsWith: info.compoundsWith };
}

/*
 * `compoundsWith` also accepts a parsed arbitrary variant, in which case the child's `compounds` is
 * computed from the selector rather than looked up. That path is `compoundsForSelectors`, which the
 * Go port reimplements; these probes pin it against the engine so the reimplementation is measured
 * rather than assumed.
 */
const compoundsForSelectorsProbes = [];
{
    const selectors = [
        '&:hover', '&:is(p)', '&::before', '&::after', '& p', '> img', '+ p', '~ p',
        '@media(width>=100px)', '@supports(display:grid)', '@container(width>=100px)',
        '@layer foo', '@page', '&:has(p)', '&:not(:hover)', 'p', '&', '&:hover::before',
    ];
    for (const parent of ['not', 'group', 'peer', 'has', 'in']) {
        for (const selector of selectors) {
            compoundsForSelectorsProbes.push({
                parent,
                selector,
                compoundsWith: designSystem.variants.compoundsWith(parent, {
                    kind: 'arbitrary',
                    selector,
                    relative: selector[0] === '>' || selector[0] === '+' || selector[0] === '~',
                }),
            });
        }
    }
}

function serializeModifier(modifier) {
    if (modifier === null || modifier === undefined) return null;
    return { kind: modifier.kind, value: modifier.value };
}

function serializeVariant(variant) {
    switch (variant.kind) {
        case 'arbitrary':
            return { kind: 'arbitrary', selector: variant.selector, relative: Boolean(variant.relative) };
        case 'static':
            return { kind: 'static', root: variant.root };
        case 'functional':
            return {
                kind: 'functional',
                root: variant.root,
                value: variant.value ? { kind: variant.value.kind, value: variant.value.value } : null,
                modifier: serializeModifier(variant.modifier),
            };
        case 'compound':
            return {
                kind: 'compound',
                root: variant.root,
                modifier: serializeModifier(variant.modifier),
                variant: serializeVariant(variant.variant),
            };
        default:
            throw new Error(`unexpected variant kind: ${variant.kind}`);
    }
}

function serializeCandidate(candidate) {
    const common = {
        variants: candidate.variants.map(serializeVariant),
        important: candidate.important,
        raw: candidate.raw,
    };
    switch (candidate.kind) {
        case 'arbitrary':
            return {
                kind: 'arbitrary',
                property: candidate.property,
                value: candidate.value,
                modifier: serializeModifier(candidate.modifier),
                ...common,
            };
        case 'static':
            return { kind: 'static', root: candidate.root, ...common };
        case 'functional':
            return {
                kind: 'functional',
                root: candidate.root,
                value: candidate.value
                    ? candidate.value.kind === 'arbitrary'
                        ? { kind: 'arbitrary', dataType: candidate.value.dataType ?? null, value: candidate.value.value }
                        : { kind: 'named', value: candidate.value.value, fraction: candidate.value.fraction ?? null }
                    : null,
                modifier: serializeModifier(candidate.modifier),
                ...common,
            };
        default:
            throw new Error(`unexpected candidate kind: ${candidate.kind}`);
    }
}

/*
 * The corpus. Hand-written classes covering each branch of the parser, then every distinct class
 * this repository actually writes, so the suite carries both the constructed edge cases and the
 * real input the exit criterion names.
 */
const syntheticClasses = [
    // Empty and degenerate.
    '', '-', '--', '!', '!!', ':', '::', 'a:', ':a', '/', 'x/', '/x',

    // Static.
    'underline', 'flex', 'sr-only', 'box-border', 'italic', 'truncate',
    'nonexistent-utility', 'group', 'peer',

    // Functional, named values.
    'bg-red-500', 'text-sm', 'p-4', 'm-0', 'w-full', 'z-10', 'gap-x-4',
    'bg-', 'bg--', 'p-', 'w-1/2', 'w-1/2/3', 'aspect-16/9', 'basis-1/3',

    // The ambiguity that decides this port. Every one of these yields more than one candidate and
    // the order is what picks the winner.
    'border-b', 'border-t', 'border-l', 'border-r', 'border-x', 'border-y',
    'border-2', 'border-b-2', 'border-b-red-500', 'border-black',
    'text-red-500', 'shadow-lg', 'ring-offset-2', 'ring-2', 'inset-x-0',
    'outline-offset-2', 'divide-x-2', 'scroll-m-4', 'space-x-4',
    'grid-cols-3', 'translate-x-4', 'bg-linear-to-r', 'font-bold',

    // Arbitrary values.
    'bg-[#0088cc]', 'bg-[color:var(--my-color)]', 'p-[10px]', 'w-[calc(100%-1rem)]',
    'bg-[]', 'bg-[ ]', 'p-[]', 'bg-[:red]', 'bg-[color:]', 'text-[12px]/[1.5]',
    'bg-[url(a_b.png)]', 'content-[\'hello_world\']', 'bg-[#0088cc]/50',
    'grid-cols-[1fr_2fr]', 'shadow-[0_0_0_1px_red]', 'bg-[rgb(0,0,0)]',
    'w-[calc(1px+2px)]', 'w-[calc(1px_+_2px)]', 'm-[-10px]',
    'bg-[var(--a,var(--b))]', 'bg-[semicolon;bad]', 'bg-[curly}bad]',
    'bg-[unbalanced(]', 'bg-[unbalanced)]', 'nonexistent-[10px]',

    // Arbitrary values in the `(…)` shorthand.
    'bg-(--my-var)', 'bg-(color:--my-var)', 'bg-(--my-var)/50', 'bg-()',
    'bg-(notavar)', 'bg-(color:notavar)', 'bg-(a:b:--c)', 'nonexistent-(--v)',

    // Arbitrary properties.
    '[color:red]', '[color:red]/50', '[color:red]!', '[--my-var:red]',
    '[-webkit-line-clamp:3]', '[Color:red]', '[0:red]', '[color]', '[:red]',
    '[color:]', '[]', '[color:red', 'color:red]', '[color:red;bad]',
    '[background:url(a_b.png)]', '[color:var(--a)]/[0.5]',

    // Modifiers, the third axis with three states.
    'bg-red-500/50', 'bg-red-500/[0.5]', 'bg-red-500/[var(--a)]', 'bg-red-500/(--a)',
    'text-sm/none', 'text-sm/6', 'text-sm/[1.5]', 'bg-red-500/50/50',
    'bg-red-500/', 'bg-red-500/[]', 'bg-red-500/()', 'bg-red-500/(notavar)',
    'bg-red-500/!', 'text-[12px]/none',

    // Importance, both spellings.
    'underline!', '!underline', 'bg-red-500!', '!bg-red-500', 'bg-red-500/50!',
    '!bg-red-500/50', 'underline!!', '!!underline', '!', 'flex!',

    // Variants: static.
    'hover:underline', 'hover:focus:underline', 'sm:flex', 'dark:bg-black',
    'first:underline', 'nonexistent:underline', 'hover:hover:underline',
    'motion-safe:underline', 'print:hidden', 'rtl:text-right',

    // Variants: functional.
    'aria-disabled:underline', 'aria-[disabled]:underline', 'data-[state=open]:flex',
    'supports-[display:grid]:flex', 'min-[400px]:flex', 'max-sm:flex',
    'nth-3:flex', 'nth-[2n+1]:flex', '@container:flex', '@lg:flex', '@max-lg:flex',
    '@[400px]:flex', '@-lg:flex', '@:flex', 'data-[]:flex', 'data-():flex',
    'data-(--a):flex', 'aria-(--a):flex', 'supports-(--a):flex',

    // Variants: compound.
    'group-hover:underline', 'peer-focus:underline', 'has-[p]:flex',
    'group-hover/name:underline', 'not-hover:underline', 'in-[p]:flex',
    'not-group-hover/name:flex', 'has-hover:flex', 'group-[&_p]:flex',
    'group-hover/foo/bar:flex', 'group-:flex', 'not-:flex', 'has-:flex',
    'group-nonexistent:flex', 'not-[&::before]:flex', 'has-[&::before]:flex',
    'group-[&::before]:flex', 'peer-[&_p]:flex', 'not-*:flex', 'has-*:flex',

    // Variants: arbitrary.
    '[&_p]:flex', '[p]:flex', '[>_img]:flex', '[+_p]:flex', '[~_p]:flex',
    '[@media(width>=100px)]:flex', '[@media(width>=100px){&:hover}]:flex',
    '[]:flex', '[_]:flex', '[&]:flex', '[@supports(display:grid)]:flex',
    '[&:hover]:[&:focus]:flex', '[&{color:red}]:flex',

    // Combinations.
    'hover:bg-red-500/50!', 'dark:hover:bg-[#0088cc]/[0.5]!',
    'group-hover/name:data-[state=open]:bg-red-500/50!',
    'sm:hover:focus:border-b-2', '!sm:hover:border-b',
    'hover:[color:red]', 'hover:[color:red]/50!',

    // Underscore decoding, which routes through decodeArbitraryValue.
    'bg-[url(https://example.com/a_b.png)]', 'font-[\'Comic_Sans\']',
    'content-[a\\_b]', 'bg-[a_b]', 'grid-cols-[repeat(2,_minmax(0,_1fr))]',
    'shadow-[inset_0_1px_--theme(--color-white/15%)]',
    'w-[calc(100%---spacing(2))]', 'bg-[theme(--color-red-500)]',
    'w-[--spacing(2)]', 'w-[var(--a_b)]', 'w-[var(--a,_1px)]',
];

const cases = [];
const seen = new Set();

/*
 * `dropWhenEmpty` is passed for the sweep only. A synthetic or repository class that reads as
 * nothing is a real assertion — the port must also read it as nothing, and `bg-[]` reading as a
 * candidate would be a defect. But the sweep crosses every root with every suffix, so most of its
 * misses are simply pairs that were never a class, and keeping 3,224 of them costs a megabyte to
 * assert what the hand-written cases already assert with intent. The sweep exists for the
 * ambiguity population, not for its misses.
 */
function record(name, source, className, dropWhenEmpty = false) {
    if (seen.has(className)) return;
    seen.add(className);
    let candidates = null;
    let threw = null;
    try {
        candidates = [...designSystem.parseCandidate(className)].map(serializeCandidate);
    } catch (error) {
        threw = String(error?.message ?? error);
    }
    if (dropWhenEmpty && threw === null && candidates.length === 0) return;
    cases.push({
        name,
        source,
        input: className,
        candidates,
        threw,
        ambiguous: candidates !== null && candidates.length > 1,
    });
}

for (const className of syntheticClasses) {
    record(className === '' ? '<empty>' : className, 'synthetic', className);
}

/*
 * Real classes. Every distinct `className="…"` token in the repository, which is the population the
 * ambiguity figure in the task body was measured against.
 */
let corpusScanned = 0;
if (corpusRootArgument) {
    const corpusRoot = NodePath.resolve(corpusRootArgument);
    const distinct = new Set();

    function scanDirectory(directory, depth) {
        if (depth > 10) return;
        let entries;
        try {
            entries = NodeFileSystem.readdirSync(directory, { withFileTypes: true });
        } catch {
            return;
        }
        for (const entry of entries) {
            if (entry.name.startsWith('.') || entry.name === 'node_modules') continue;
            const fullPath = NodePath.join(directory, entry.name);
            if (entry.isDirectory()) {
                scanDirectory(fullPath, depth + 1);
                continue;
            }
            if (!entry.name.endsWith('.tsx') && !entry.name.endsWith('.ts')) continue;
            let contents;
            try {
                contents = NodeFileSystem.readFileSync(fullPath, 'utf8');
            } catch {
                continue;
            }
            for (const match of contents.matchAll(/className="([^"]*)"/g)) {
                for (const className of match[1].split(/\s+/)) {
                    if (className.length === 0) continue;
                    corpusScanned++;
                    distinct.add(className);
                }
            }
        }
    }
    scanDirectory(corpusRoot, 0);

    for (const className of [...distinct].sort()) {
        record(className, 'repository', className);
    }
}

/*
 * An exhaustive sweep over the ambiguity mechanism itself. Every utility root crossed with the
 * value shapes that make a root split more than one way, so the ordering claim rests on a
 * population rather than on the handful of hand-picked classes above.
 */
{
    const roots = [...designSystem.utilities.utilities.keys()];
    const suffixes = ['b', 't', 'l', 'r', 'x', 'y', '2', 'red-500', 'offset-2', 'full', 'auto', '0', 'sm'];

    /*
     * The sweep is recorded into a staging list and filtered before it joins the corpus. Every
     * multi-candidate reading is kept, because those are the population the ordering claim rests
     * on and there is no such thing as a redundant one. Single-candidate readings are sampled: they
     * assert the same unambiguous split the hand-written and repository cases already assert with
     * intent, and keeping all 8,685 of them costs two and a half megabytes of fixture to restate
     * it. Ambiguity is the signal here; a root that splits exactly one way is the background.
     */
    const staged = [];
    const stagedSeen = new Set();
    for (const root of roots) {
        if (root.indexOf('-') === -1) continue;
        for (const suffix of suffixes) {
            const className = `${root}-${suffix}`;
            if (seen.has(className) || stagedSeen.has(className)) continue;
            stagedSeen.add(className);
            let candidates;
            try {
                candidates = [...designSystem.parseCandidate(className)].map(serializeCandidate);
            } catch {
                continue;
            }
            if (candidates.length === 0) continue;
            staged.push({ className, candidates });
        }
    }

    let unambiguousKept = 0;
    for (const entry of staged) {
        if (entry.candidates.length === 1) {
            // Deterministic sample rather than random, so regenerating produces the same fixture
            // and `-check` means something.
            if (unambiguousKept % 8 !== 0) {
                unambiguousKept++;
                continue;
            }
            unambiguousKept++;
        }
        seen.add(entry.className);
        cases.push({
            name: entry.className,
            source: 'sweep',
            input: entry.className,
            candidates: entry.candidates,
            threw: null,
            ambiguous: entry.candidates.length > 1,
        });
    }
}

const ambiguousCount = cases.filter((testCase) => testCase.ambiguous).length;
const repositoryCases = cases.filter((testCase) => testCase.source === 'repository');
const ambiguousRepositoryCount = repositoryCases.filter((testCase) => testCase.ambiguous).length;

/*
 * Volume assertions. A fixture that captured almost nothing looks exactly like a fixture that
 * captured everything and agreed, so the counts that make the suite meaningful are asserted here
 * rather than hoped for.
 */
const failures = [];

/*
 * The root counts alone cannot tell whether this repository's stylesheets loaded. Tailwind registers
 * its ~1,201 built-in utility roots from JavaScript, not from CSS, so pointing this tool at a
 * stylesheet that imports nothing still reports 1,201 roots and 83 variants. That was measured, not
 * assumed: a control run against an empty theme passed a `>= 500` assertion comfortably, which is
 * how this check was found to be worthless in its first form.
 *
 * What actually discriminates is the delta. The repository's own `@utility` blocks and `@theme`
 * breakpoints contribute roots the framework alone never registers, so a handful of them are named
 * here as markers. If the theme graph fails to load, these disappear and the capture stops rather
 * than quietly describing framework defaults while claiming to describe this repository.
 */
const repositoryUtilityMarkers = ['scrollbar-hide', 'prose', 'fade-in', 'slide-in-from-top'];
const repositoryVariantMarkers = ['sm', 'md', 'lg', 'xl', '2xl'];

if (Object.keys(utilityRoots).length < 500) {
    failures.push(`only ${Object.keys(utilityRoots).length} utility roots; even the framework defaults are missing`);
}
if (Object.keys(variantRoots).length < 50) {
    failures.push(`only ${Object.keys(variantRoots).length} variant roots; even the framework defaults are missing`);
}
if (themeEntry) {
    const missingUtilities = repositoryUtilityMarkers.filter((marker) => !(marker in utilityRoots));
    if (missingUtilities.length > 0) {
        failures.push(
            `the repository's own @utility roots are absent (${missingUtilities.join(', ')}); ` +
                'the theme entry point was named but its @import graph did not load, so these tables are framework defaults',
        );
    }
    const missingVariants = repositoryVariantMarkers.filter((marker) => !(marker in variantRoots));
    if (missingVariants.length > 0) {
        failures.push(
            `the repository's own breakpoint variants are absent (${missingVariants.join(', ')}); ` +
                'the @theme block did not load',
        );
    }
}
if (ambiguousCount < 100) {
    failures.push(`only ${ambiguousCount} ambiguous classes; the ordering claim rests on too few`);
}
if (corpusRootArgument && repositoryCases.length < 100) {
    failures.push(`only ${repositoryCases.length} repository classes; the corpus scan found the wrong directory`);
}
if (failures.length > 0) {
    for (const failure of failures) process.stderr.write(`${failure}\n`);
    process.exit(4);
}

process.stdout.write(
    JSON.stringify(
        {
            tailwindVersion: packageJson.version,
            parseCandidateResolvedByIdentity: true,
            themeEntry: themeEntry ? NodePath.basename(themeEntry) : null,
            utilityRootCount: Object.keys(utilityRoots).length,
            variantRootCount: Object.keys(variantRoots).length,
            repositoryUtilityMarkers,
            repositoryVariantMarkers,
            classNameOccurrences: corpusScanned,
            repositoryClassCount: repositoryCases.length,
            ambiguousCount,
            ambiguousRepositoryCount,
            utilityRoots,
            variantRoots,
            compoundsForSelectorsProbes,
            cases,
        },
        null,
        2,
    ) + '\n',
);
