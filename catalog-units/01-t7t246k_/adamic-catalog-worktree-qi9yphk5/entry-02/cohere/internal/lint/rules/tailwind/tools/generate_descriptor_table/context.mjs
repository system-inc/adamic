/*
 * The parts of the descriptor table that are not readings.
 *
 * `extract.mjs` measures readings and is the tool that proved the model; it is left untouched. Four
 * things the Go lookup needs are not readings and so are not in its output, and they are collected
 * here from the same design system by the same loader:
 *
 *   - `namespaces`, longest first. The order is the precedence the engine applies and it is part of
 *     the contract rather than a presentation choice: `--drop-shadow` must be tested before
 *     `--shadow`, or `drop-shadow-lg` resolves through the wrong one.
 *
 *   - `keysByNamespace`, which is what decides whether a bare value is a theme key at all. This is
 *     the per-repository half of the table, and it is why the table is generated rather than
 *     committed.
 *
 *   - `propertyOrder`, needed only to turn an arbitrary property such as `[font:inherit]` into an
 *     index. It is read out of the bundle as a plain string array rather than derived by asking the
 *     engine to sort one utility per property, because that route only reaches properties some
 *     utility declares alone: the same comment in `generate_collapse/enumerate.mjs` records it
 *     finding 254 of 359 and shifting every index after a missing one.
 *
 *   - `perDeclarationRoots`, the roots whose arity is the count of surviving `@utility` declarations
 *     rather than a function of any data type. These are the 18 known registry exceptions, and the
 *     table has to know which roots it cannot answer so the lookup can decline instead of guessing.
 *
 * Usage:
 *
 *   node internal/lint/rules/tailwind/tools/generate_descriptor_table/context.mjs <theme.css> [--resolve-root <dir>] > context.json
 */

import NodeFileSystem from 'node:fs';
import NodeModule from 'node:module';
import NodePath from 'node:path';
import { loadDesignSystem, parseCandidate, readingOf } from '../generate_descriptors/loader.mjs';

const entryPointArgument = process.argv[2];
if (!entryPointArgument) {
    process.stderr.write('usage: context.mjs <theme.css> [--resolve-root <dir>]\n');
    process.exit(2);
}

/*
 * `--resolve-root` separates where the CSS lives from where bare specifiers resolve, matching
 * `extract.mjs`. The two tools are a pair generated from one stylesheet and consumed together, so a
 * design system only one of them could open would be half-generable, which is what
 * `testdata/independent_theme.css` was.
 */
const resolveRootFlagIndex = process.argv.indexOf('--resolve-root');
const resolveRootArgument = resolveRootFlagIndex >= 0 ? process.argv[resolveRootFlagIndex + 1] : null;

const { designSystem, tailwindVersion, entryPoint, recordedEntryPoint, resolveRoot } = await loadDesignSystem(
    entryPointArgument,
    resolveRootArgument,
);

/*
 * The namespaces, confirmed against the design system's own membership test.
 *
 * A key is `--<namespace>-<name>` and the boundary is ambiguous from the string alone, since
 * `--color-red-500` could split three ways. Every candidate prefix is therefore offered to
 * `keysInNamespaces`, which is the engine's answer rather than a guess about where the boundary is.
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

// Longest first, then lexical, which is the order `extract.mjs` probes in and therefore the order
// its readings were measured under. Sorting these two differently would make the table's readings
// answer a different question than the lookup asks.
const namespaces = Object.keys(keysByNamespace).sort(
    (left, right) => right.length - left.length || (left < right ? -1 : 1),
);

/*
 * Tailwind's global property order, read out of the installed bundle as a plain string array.
 */
const propertyOrder = [];
{
    let bundleSource = '';
    try {
        const requireFromResolveRoot = NodeModule.createRequire(NodePath.join(resolveRoot, 'noop.js'));
        const tailwindRoot = NodePath.dirname(requireFromResolveRoot.resolve('tailwindcss/package.json'));
        bundleSource = NodeFileSystem.readFileSync(NodePath.join(tailwindRoot, 'dist', 'lib.mjs'), 'utf8');
    }
    catch {
        bundleSource = '';
    }
    const orderMatch = bundleSource.match(/\["container-type","pointer-events"[^\]]*\]/);
    if (orderMatch === null) {
        // Loud rather than empty. An empty property order silently reads every arbitrary property as
        // unknown, which sorts them all together and looks like a stable answer.
        process.stderr.write('context.mjs: could not locate the property order in the bundle\n');
        process.exit(3);
    }
    for (const property of JSON.parse(orderMatch[0])) propertyOrder.push(property);
}

/*
 * The per-declaration roots, found by measurement rather than by a hardcoded list.
 *
 * The mechanism is that a repository `@utility` block can hold several declarations that each
 * resolve independently, so arity is the count of the ones that survive and a value can satisfy more
 * than one path at once. `slide-in-from-top-50` survives two declarations because `50` is both an
 * `integer` and a `--percentage-*` key; `slide-in-from-top-4` survives one.
 *
 * That is exactly what makes them detectable: probe each root with a value that is only an integer
 * and with one that is also a theme key, and a root whose count changes between them is resolving
 * per declaration. This finds the roots rather than naming them, so a repository that adds another
 * `@utility` of the same shape is covered without editing this file, and a repository that has none
 * reports none.
 */
const perDeclarationRoots = [];
{
    const registryClassNames = [];
    for (const entry of designSystem.getClassList()) {
        registryClassNames.push(Array.isArray(entry) ? entry[0] : entry);
    }

    const functionalRoots = new Set();
    for (const className of registryClassNames) {
        for (const probe of [className, className + '-4']) {
            const candidate = parseCandidate(designSystem, probe);
            if (candidate?.kind === 'functional' && candidate.root) functionalRoots.add(candidate.root);
        }
    }

    // The values that separate the two resolution paths: one only an integer, and the theme keys of
    // every `--percentage*` namespace, which are what a value can additionally satisfy.
    const percentageKeys = new Set();
    for (const namespace of namespaces) {
        if (!namespace.startsWith('--percentage')) continue;
        for (const key of keysByNamespace[namespace]) percentageKeys.add(key);
    }

    for (const root of Array.from(functionalRoots).sort()) {
        const integerOnly = readingOf(designSystem, root + '-4');
        if (integerOnly === null) continue;
        for (const key of percentageKeys) {
            const both = readingOf(designSystem, root + '-' + key);
            if (both === null) continue;
            if (both.count !== integerOnly.count) {
                perDeclarationRoots.push(root);
                break;
            }
        }
    }
}

/*
 * Static utilities the engine accepts and `getClassList()` does not advertise.
 *
 * `extract.mjs` builds its static list from the registry, which is the right source for a
 * measurement of the registry. It is not a complete list of what the engine will compile: probing
 * every functional root with a handful of keywords finds `order-none`, which parses as a static
 * utility rooted at `order-none`, compiles to `[17]#1`, and appears nowhere in `getClassList()`.
 *
 * One class, on this design system, and it is recovered rather than hardcoded because the point is
 * that the registry is not exhaustive. A lint rule reads what an author wrote, and an author can
 * write a class the registry never listed; the table declining it would be a wrong answer to a real
 * class. The keyword list is the set of value-like words that also name static utilities, which is
 * where the gap can occur at all.
 */
const unlistedStatics = {};
{
    const registry = new Set();
    for (const entry of designSystem.getClassList()) {
        registry.add(Array.isArray(entry) ? entry[0] : entry);
    }

    const functionalRoots = new Set();
    for (const className of registry) {
        for (const probe of [className, className + '-4']) {
            const candidate = parseCandidate(designSystem, probe);
            if (candidate?.kind === 'functional' && candidate.root) functionalRoots.add(candidate.root);
        }
    }

    for (const root of Array.from(functionalRoots).sort()) {
        for (const keyword of ['none', 'auto', 'full', 'first', 'last', 'initial', 'inherit']) {
            const probe = root + '-' + keyword;
            if (registry.has(probe)) continue;
            const candidate = parseCandidate(designSystem, probe);
            if (candidate?.kind !== 'static' || candidate.root !== probe) continue;
            const reading = readingOf(designSystem, probe);
            if (reading === null) continue;
            unlistedStatics[probe] = { order: reading.order, count: reading.count };
        }
    }
}

process.stdout.write(JSON.stringify({
    tailwindVersion,
    entryPoint: recordedEntryPoint,
    namespaces,
    keysByNamespace,
    propertyOrder,
    perDeclarationRoots,
    unlistedStatics,
}, null, 2) + '\n');
