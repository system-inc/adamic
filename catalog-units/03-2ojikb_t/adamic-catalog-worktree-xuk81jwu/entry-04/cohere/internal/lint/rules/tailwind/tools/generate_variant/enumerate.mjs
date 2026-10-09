/*
 * Ask the real Tailwind 4.3.3 engine how it orders variants, and emit both the registry it built
 * and the order it produced for a corpus of class lists.
 *
 * The Go port of the variant half of the sort is checked against these answers rather than against
 * a reading of `variants.ts`, for the reason generate_theme exists: the two are different
 * claims. The one that decides this component is `getVariantOrder`. Reading `variants.ts` suggests
 * a variant has a fixed position, and cohere has shipped a 145-entry table of exactly that shape.
 * Asking the engine says the position is assigned per run, densely, over only the variants that
 * were actually parsed, with ties collapsed to a shared index. `hover` is 0 in a run that parsed
 * two variants and 1 in a run that parsed six. The absolute number is not a fact about `hover`.
 *
 * # Why this tool asks the design system rather than the Variants class
 *
 * `Variants` is not an export of the published package. Like `Theme` it is reachable through
 * `__unstable__loadDesignSystem`, whose returned design system exposes the live registry as
 * `variants` and the two methods under test as `getVariantOrder` and `getClassOrder`.
 *
 * `getClassOrder` is the better ground truth of the two, and it is what this tool leans on. It is
 * the engine's whole sort, run over a real class list, returning each class's final rank. It is
 * also the public API Tailwind's own Prettier plugin consumes, so a Go port that agrees with it
 * agrees with the thing the ecosystem treats as the answer. Comparing a rendered selector would
 * have tested the 1,000 lines of `variants.ts` that cohere deliberately does not port; comparing
 * the rank tests exactly the part it does.
 *
 * # The three corpora, and why all three
 *
 * `registry` captures every registered variant root with the order number and kind the engine gave
 * it, plus which orders carry a comparison function. That is the static half: 88 roots, four
 * shared-order groups, four compare functions.
 *
 * `repository` cases load a real repository's stylesheet and sort that repository's own class
 * lists. This is the exit criterion. Both repositories in the corpus declare `@custom-variant dark`
 * with a `:has()` selector, so a port that knows only framework variants gets the framework's
 * `dark` and is wrong here in exactly the per-repository way this port exists to fix.
 *
 * `synthetic` cases are hand-built class lists that reach the pairs a real corpus never writes:
 * a stacked variant against a single one, two arbitrary variants against each other, breakpoints
 * out of registration order, container queries, and named groups. A real corpus exercises the
 * common half; the other half is where a port fails silently.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodePath from 'node:path';
import { locate } from '../corpus.mjs';

const [, , packageRootArgument, corpusPathArgument] = process.argv;
if (!packageRootArgument) {
    process.stderr.write('usage: enumerate.mjs <tailwindcss package root> [corpus.json]\n');
    process.exit(2);
}

const packageRoot = NodePath.resolve(packageRootArgument);
const packageJson = JSON.parse(NodeFileSystem.readFileSync(NodePath.join(packageRoot, 'package.json'), 'utf8'));
const bundle = await import(NodePath.join(packageRoot, 'dist', 'lib.mjs'));

/*
 * Resolve `@import` the way a bundler would, matching generate_theme's loader exactly. The two
 * tools read the same repository stylesheets and a divergence between their loaders would show up
 * as a theme difference rather than as the loader bug it is.
 */
function loadStylesheet(identifier, base) {
    let resolved;
    if (identifier === 'tailwindcss') {
        resolved = NodePath.join(packageRoot, 'index.css');
    } else if (identifier.startsWith('tailwindcss/')) {
        resolved = NodePath.join(packageRoot, identifier.slice('tailwindcss/'.length));
    } else {
        resolved = NodePath.resolve(base, identifier);
    }
    if (!resolved.endsWith('.css')) resolved += '.css';
    return {
        base: NodePath.dirname(resolved),
        content: NodeFileSystem.readFileSync(resolved, 'utf8'),
    };
}

/*
 * Stub `@plugin` and `@config`, matching generate_theme's loader.
 *
 * A JavaScript plugin can register variants, so stubbing one is a real narrowing and it is stated
 * rather than hidden: a plugin that called `addVariant` would contribute a root this fixture does
 * not see. Neither repository in the corpus registers a variant that way. Executing arbitrary
 * repository plugins to build a fixture is the larger risk, and the narrowing is visible here
 * where a later reader will find it.
 */
function loadModule(identifier, base) {
    return { base, path: identifier, module: {} };
}

async function loadDesignSystem(css, base) {
    return await bundle.__unstable__loadDesignSystem(css, {
        base,
        loadStylesheet: async (identifier, importBase) => loadStylesheet(identifier, importBase),
        loadModule: async (identifier, importBase) => loadModule(identifier, importBase),
    });
}

/*
 * Describe a parsed variant as plain data the Go side can compare against.
 *
 * The shape mirrors `Variant` in `candidate.ts`: a static variant is a root, a functional one adds
 * a value and a modifier, a compound one nests another variant, and an arbitrary one carries a
 * selector instead of a root. The nesting is recursive because `group-[&:hover]:` is a compound
 * wrapping an arbitrary, and flattening it would lose the inner kind that decides the comparison.
 */
function describeVariant(variant) {
    if (variant === null || variant === undefined) return null;
    switch (variant.kind) {
        case 'arbitrary':
            return { kind: 'arbitrary', selector: variant.selector, relative: variant.relative ?? false };
        case 'static':
            return { kind: 'static', root: variant.root };
        case 'functional':
            return {
                kind: 'functional',
                root: variant.root,
                value: variant.value === null ? null : { kind: variant.value.kind, value: variant.value.value },
                modifier: variant.modifier === null ? null : { kind: variant.modifier.kind, value: variant.modifier.value },
            };
        case 'compound':
            return {
                kind: 'compound',
                root: variant.root,
                modifier: variant.modifier === null ? null : { kind: variant.modifier.kind, value: variant.modifier.value },
                variant: describeVariant(variant.variant),
            };
        default:
            return { kind: variant.kind };
    }
}

/*
 * Capture the engine's registry: every registered root, the order number it holds, and its kind.
 *
 * The order numbers are captured rather than derived because the shared ones are the finding.
 * `sm`, `md`, `lg`, `xl`, `2xl` and `min` all hold order 64 in a default build, which is what
 * `variants.group` does, and it means the five breakpoints do not order each other by registration
 * at all. They order by the compare function registered against order 64, which resolves them
 * through the theme's `--breakpoint` namespace. A port that gave each breakpoint its own position
 * agrees on a default theme and diverges on any repository that redefines one.
 */
function describeRegistry(designSystem) {
    const entries = [];
    for (const [name, info] of designSystem.variants.entries()) {
        entries.push({ name, order: info.order, kind: info.kind });
    }
    return {
        entries,
        compareFnOrders: Array.from(designSystem.variants.compareFns.keys()).sort((a, z) => a - z),
    };
}

/*
 * Capture `getVariantOrder` for a set of parsed variants.
 *
 * The design system is fresh per call, because the order map is a function of exactly which
 * variants have been parsed into the design system's cache. Reusing one across cases would let an
 * earlier case's variants widen a later case's map, which is the same defect as the run-dependence
 * this fixture exists to record, hidden inside the tool that records it.
 */
function describeVariantOrder(designSystem, variantStrings) {
    const parsed = [];
    for (const variantString of variantStrings) {
        const variant = designSystem.parseVariant(variantString);
        parsed.push({ raw: variantString, variant: describeVariant(variant) });
    }

    const order = designSystem.getVariantOrder();
    const positions = [];
    for (const entry of parsed) {
        if (entry.variant === null) {
            positions.push({ raw: entry.raw, variant: null, index: null });
            continue;
        }
        const variantObject = designSystem.parseVariant(entry.raw);
        const index = order.get(variantObject);
        positions.push({
            raw: entry.raw,
            variant: entry.variant,
            index: index === undefined ? null : index,
        });
    }

    return { size: order.size, positions };
}

/*
 * Capture `getClassOrder` for a class list, plus the per-candidate bitmask that produced it.
 *
 * Both are recorded because they answer different questions. The rank is the observable behaviour
 * and the thing a consumer sorts on. The bitmask is the mechanism, and recording it is what turns
 * a passing test into evidence about *why* it passes: two ports can agree on every rank in a
 * corpus while one of them builds the mask by OR-ing bits and the other by comparing segments, and
 * only the mask distinguishes them.
 *
 * A class the engine cannot place comes back as `null` from `getClassOrder`, and that is kept
 * rather than dropped, because "this is not a Tailwind class" is an answer the port must reproduce.
 */
function describeClassOrder(designSystem, classes) {
    const ranks = designSystem.getClassOrder(classes);

    const variantOrderMap = designSystem.getVariantOrder();
    const masks = [];
    for (const className of classes) {
        const candidates = Array.from(designSystem.parseCandidate(className));
        if (candidates.length === 0) {
            masks.push({ className, mask: null, variants: null });
            continue;
        }
        // The first candidate is the one `compileCandidates` sorts on when several parse, matching
        // `getClassOrder`'s own "take the position of the first one".
        const candidate = candidates[0];
        let mask = 0n;
        let resolvable = true;
        const described = [];
        for (const variant of candidate.variants) {
            const index = variantOrderMap.get(variant);
            described.push({ variant: describeVariant(variant), index: index === undefined ? null : index });
            if (index === undefined) {
                resolvable = false;
                continue;
            }
            mask |= 1n << BigInt(index);
        }
        masks.push({
            className,
            mask: resolvable ? mask.toString(10) : null,
            variants: described,
        });
    }

    return {
        ranks: ranks.map(([className, rank]) => ({ className, rank: rank === null ? null : rank.toString(10) })),
        masks,
        sorted: ranks
            .filter(([, rank]) => rank !== null)
            .slice()
            .sort((a, z) => (a[1] < z[1] ? -1 : a[1] > z[1] ? 1 : 0))
            .map(([className]) => className),
    };
}

/*
 * Synthetic class lists, each chosen for a pair a real corpus does not write.
 *
 * The first group is the one this whole component exists to settle: a stacked variant against a
 * single one. cohere's shipped comparator holds that every single variant precedes every stacked
 * one, arrived at by fixture failure. If the engine agrees, these lists confirm it on pairs the
 * original four class lists never covered. If it disagrees, these are where it shows.
 */
const syntheticClassLists = [
    { name: 'stacked-against-single', classes: ['dark:flex', 'group-hover:disabled:flex'] },
    { name: 'stacked-between-singles', classes: ['dark:flex', 'group-hover:flex', 'group-hover:disabled:flex', 'hover:flex'] },
    { name: 'depth-first-probe', classes: ['hover:flex', 'dark:hover:flex', 'dark:flex', 'focus:flex', 'dark:focus:flex'] },
    { name: 'compound-inner-segment', classes: ['dark:placeholder:flex', 'dark:focus:flex'] },
    { name: 'bare-before-variant', classes: ['hover:px-8', 'px-4'] },
    { name: 'breakpoints-in-order', classes: ['sm:flex', 'md:flex', 'lg:flex', 'xl:flex', '2xl:flex'] },
    { name: 'breakpoints-shuffled', classes: ['lg:flex', '2xl:flex', 'sm:flex', 'xl:flex', 'md:flex'] },
    { name: 'breakpoints-against-state', classes: ['sm:flex', 'hover:flex', 'dark:flex', 'md:flex'] },
    { name: 'min-max-breakpoints', classes: ['max-sm:flex', 'min-sm:flex', 'sm:flex', 'max-lg:flex'] },
    { name: 'arbitrary-breakpoints', classes: ['min-[320px]:flex', 'min-[640px]:flex', 'max-[480px]:flex'] },
    { name: 'container-queries', classes: ['@sm:flex', '@lg:flex', '@max-sm:flex', '@min-lg:flex'] },
    { name: 'container-query-modifiers', classes: ['@sm/main:flex', '@sm/side:flex', '@lg/main:flex'] },
    { name: 'arbitrary-variant', classes: ['[&:nth-child(3)]:flex', 'hover:flex', 'dark:flex'] },
    { name: 'two-arbitrary-variants', classes: ['[&:nth-child(3)]:flex', '[&:nth-child(2)]:flex'] },
    { name: 'arbitrary-against-stacked', classes: ['[&:hover]:flex', 'dark:hover:flex', 'hover:flex'] },
    { name: 'named-groups', classes: ['group-hover/csv:flex', 'group-hover/pdf:flex', 'group-hover:flex'] },
    { name: 'named-peers', classes: ['peer-focus/one:flex', 'peer-checked/two:flex', 'peer-hover:flex'] },
    { name: 'functional-data', classes: ['data-[show=false]:flex', 'data-[show=true]:flex', 'hover:flex'] },
    { name: 'functional-aria', classes: ['aria-expanded:flex', 'aria-[sort=ascending]:flex', 'hover:flex'] },
    { name: 'nth-variants', classes: ['nth-3:flex', 'nth-last-2:flex', 'nth-of-type-2:flex'] },
    { name: 'supports-variant', classes: ['supports-[display:grid]:flex', 'hover:flex'] },
    { name: 'not-variant', classes: ['not-hover:flex', 'hover:flex', 'not-dark:flex'] },
    { name: 'has-variant', classes: ['has-checked:flex', 'has-[>img]:flex', 'hover:flex'] },
    { name: 'in-variant', classes: ['in-hover:flex', 'in-focus:flex', 'hover:flex'] },
    { name: 'pseudo-elements', classes: ['before:flex', 'after:flex', 'placeholder:flex', 'marker:flex'] },
    { name: 'triple-stack', classes: ['dark:hover:focus:flex', 'dark:hover:flex', 'dark:flex', 'flex'] },
    { name: 'stack-order-independence', classes: ['dark:hover:flex', 'hover:dark:flex'] },
    { name: 'unknown-variant', classes: ['nonsense-variant:flex', 'hover:flex'] },
    { name: 'important-and-variant', classes: ['hover:flex!', 'hover:flex', 'flex!'] },
    { name: 'motion-and-contrast', classes: ['motion-safe:flex', 'motion-reduce:flex', 'contrast-more:flex'] },
    { name: 'direction-variants', classes: ['ltr:flex', 'rtl:flex', 'dark:flex'] },
    { name: 'print-and-media', classes: ['print:flex', 'portrait:flex', 'landscape:flex', 'dark:flex'] },
    { name: 'star-variants', classes: ['*:flex', '**:flex', 'hover:flex'] },
    { name: 'peer-and-group-mixed', classes: ['peer-hover:flex', 'group-hover:flex', 'hover:flex'] },
];

/*
 * Synthetic variant-order cases, which record `getVariantOrder` directly rather than through a
 * class list.
 *
 * These are what make the run-dependence visible. The same variant string appears in cases of
 * different sizes, and the fixture records a different index for it in each, which is the fact a
 * static table cannot represent.
 */
const syntheticVariantOrderCases = [
    { name: 'two-variants', variants: ['hover', 'focus'] },
    { name: 'six-variants', variants: ['dark', 'hover', 'focus', 'sm', 'md', 'group-hover'] },
    { name: 'hover-alone', variants: ['hover'] },
    { name: 'breakpoint-tie-group', variants: ['sm', 'md', 'lg', 'xl', '2xl'] },
    { name: 'breakpoints-reversed-input', variants: ['2xl', 'xl', 'lg', 'md', 'sm'] },
    { name: 'min-and-max', variants: ['min-sm', 'max-sm', 'min-lg', 'max-lg'] },
    { name: 'compound-and-inner', variants: ['group-hover', 'hover', 'group-focus'] },
    { name: 'arbitrary-variants', variants: ['[&:hover]', '[&:focus]'] },
    { name: 'container-queries', variants: ['@sm', '@lg', '@max-sm'] },
    { name: 'modifiers', variants: ['group-hover/one', 'group-hover/two', 'group-hover'] },
];

/*
 * Stylesheets that register `@custom-variant`, which is the per-repository half of this component.
 *
 * The two cases are deliberately different in kind, because the engine treats them differently and
 * the difference is not what the name suggests. A `@custom-variant` under a NEW name is appended
 * past the framework's last order number and genuinely shifts what a repository's classes sort
 * against. A `@custom-variant` that reuses a framework name overwrites the registration in place:
 * `Variants.set` does `Object.assign(existing, { kind, applyFn, compounds })` and never touches
 * `order`, so the selector changes completely while the sort position does not move at all.
 *
 * Both repositories in the corpus do the second kind. Their emitted `dark:` selector is their own
 * `:has()` rule rather than the framework's `@media (prefers-color-scheme: dark)`, and their
 * variant ordering is nevertheless identical to the framework's. A port that assumed a custom
 * variant always means a new position would be wrong on exactly the case both repositories write.
 */
const customVariantStylesheets = [
    {
        name: 'custom-variant-new-name',
        css: '@import "tailwindcss";\n@custom-variant sidebar (&:where(.sidebar *));',
        classes: ['dark:flex', 'sidebar:flex', 'hover:flex'],
    },
    {
        name: 'custom-variant-overrides-framework',
        css: '@import "tailwindcss";\n@custom-variant dark (&:where(.scheme-dark, .scheme-dark *));',
        classes: ['dark:flex', 'hover:flex', 'print:flex'],
    },
    {
        name: 'custom-variant-two-new-names',
        css: '@import "tailwindcss";\n@custom-variant sidebar (&:where(.sidebar *));\n@custom-variant panel (&:where(.panel *));',
        classes: ['panel:flex', 'sidebar:flex', 'dark:flex'],
    },
];

const cases = [];

/*
 * The default design system: framework variants only, no repository stylesheet.
 *
 * This is the baseline the registry is captured from, and it is deliberately separate from the
 * repository cases. A repository that registers `@custom-variant dark` overwrites the framework's
 * `dark` in place, keeping its order number, and that overwrite is only visible as a difference
 * against a build that did not have it.
 */
{
    const designSystem = await loadDesignSystem('@import "tailwindcss";', '/');
    cases.push({
        name: 'framework-default',
        source: 'default',
        input: '@import "tailwindcss";',
        entryPath: '',
        registry: describeRegistry(designSystem),
    });

    for (const syntheticCase of syntheticVariantOrderCases) {
        const fresh = await loadDesignSystem('@import "tailwindcss";', '/');
        cases.push({
            name: `variant-order/${syntheticCase.name}`,
            source: 'synthetic',
            input: '@import "tailwindcss";',
            entryPath: '',
            variantOrder: describeVariantOrder(fresh, syntheticCase.variants),
        });
    }

    for (const syntheticCase of syntheticClassLists) {
        const fresh = await loadDesignSystem('@import "tailwindcss";', '/');
        cases.push({
            name: `class-order/${syntheticCase.name}`,
            source: 'synthetic',
            input: '@import "tailwindcss";',
            entryPath: '',
            classOrder: describeClassOrder(fresh, syntheticCase.classes),
        });
    }
}

/*
 * The `@custom-variant` stylesheets, captured with both a registry and a class order.
 *
 * The registry is what shows the order number a new name received; the class order is what shows
 * whether that number changed any observable ordering.
 */
for (const customCase of customVariantStylesheets) {
    const registryDesignSystem = await loadDesignSystem(customCase.css, '/');
    const orderDesignSystem = await loadDesignSystem(customCase.css, '/');
    cases.push({
        name: `custom-variant/${customCase.name}`,
        source: 'synthetic',
        input: customCase.css,
        entryPath: '',
        registry: describeRegistry(registryDesignSystem),
        classOrder: describeClassOrder(orderDesignSystem, customCase.classes),
    });
}

/*
 * Repository cases: a real stylesheet, and that repository's own class literals.
 *
 * The literals arrive from the Go side, which reads them out of the repository's source tree. That
 * split is deliberate: Go already has the class-literal reader this repository's rules use, and
 * reimplementing it in JavaScript would mean the fixture and the rule disagreed about what a class
 * literal is, which is a difference that would read as an ordering divergence.
 */
if (corpusPathArgument) {
    const corpus = JSON.parse(NodeFileSystem.readFileSync(NodePath.resolve(corpusPathArgument), 'utf8'));
    for (const entry of corpus) {
        // A corpus spelling is recorded as given, so the fixture reads the same on every machine.
        const entryLocated = locate(entry.path);
        const css = NodeFileSystem.readFileSync(entryLocated.path, 'utf8');
        const base = NodePath.dirname(entryLocated.path);

        const registryDesignSystem = await loadDesignSystem(css, base);
        cases.push({
            name: entry.name,
            source: 'repository',
            input: '',
            entryPath: entryLocated.recorded,
            registry: describeRegistry(registryDesignSystem),
        });

        const classLists = entry.classLists ?? [];
        const captured = [];
        for (let index = 0; index < classLists.length; index++) {
            const fresh = await loadDesignSystem(css, base);
            captured.push({
                name: `${entry.name}/list-${index}`,
                classes: classLists[index],
                classOrder: describeClassOrder(fresh, classLists[index]),
            });
        }
        cases.push({
            name: `${entry.name}/class-lists`,
            source: 'repository',
            input: '',
            entryPath: entryLocated.recorded,
            classLists: captured,
        });
    }
}

const syntheticCaseCount = cases.filter((aCase) => aCase.source === 'synthetic').length;
const repositoryCaseCount = cases.filter((aCase) => aCase.source === 'repository').length;

process.stdout.write(
    JSON.stringify(
        {
            tailwindVersion: packageJson.version,
            syntheticCaseCount,
            repositoryCaseCount,
            cases,
        },
        null,
        4,
    ) + '\n',
);
