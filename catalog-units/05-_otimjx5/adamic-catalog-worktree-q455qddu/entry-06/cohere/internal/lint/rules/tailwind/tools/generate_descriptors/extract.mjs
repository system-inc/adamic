/*
 * Extract the descriptor table, and prove it against the engine.
 *
 * The architecture of the Go port rests on one claim: a utility root's reading — `{order, count}`,
 * the only thing `getPropertySort` computes — is a pure function of the inferred data type of its
 * value. If that holds, `utilities.ts` (6,827 lines, most of it computing values nobody reads)
 * collapses to a table of a few hundred rows. If it does not, the shape of the port changes.
 *
 * This tool tests the claim before any Go is written, in three passes:
 *
 *   Extract   For every functional root, probe it with every shape, read the engine's answer, and
 *             ask whether the shape's inferred data type determines that answer. The output is
 *             `{root -> {typeList, readingByType, fallback}}` plus every root where it does not.
 *
 *   Replay    Take the extracted descriptors and predict the reading of all ~37,600 classes in the
 *             registry. The registry is the population; a corpus is a sample of what someone
 *             happened to write, and an earlier table built from a corpus missed nine families.
 *
 *   Sweep     Predict the reading of every root x every shape, ~166,000 probes, which is the part
 *             that reaches values no registry contains.
 *
 * Three controls, because a sweep that predicts everything is exactly what a broken sweep looks
 * like:
 *
 *   A planted dirty descriptor, which the harness must flag. A check that has never returned a
 *   positive has not been shown to be able to.
 *
 *   Volume assertions. 37,643 classes and six figures of probes are the expected magnitudes; a run
 *   reporting far fewer measured the wrong thing and its clean result means nothing.
 *
 *   A null count, reported separately and never scored as agreement. In the ahra corpus 10 of 1,208
 *   distinct classes return no reading at all, and counting those as passes would be blindest on
 *   exactly the classes most likely to break.
 *
 * Usage:
 *
 *   node internal/lint/rules/tailwind/tools/generate_descriptors/extract.mjs <theme.css> [--json <path>]
 */

import * as NodeFileSystem from 'node:fs';
import * as NodeModule from 'node:module';
import * as NodePath from 'node:path';
import { loadDesignSystem, parseCandidate, readingKey, readingOf, surveyOwnContributions } from './loader.mjs';
import { DataTypeNames, probeClassName, Shapes } from './shapes.mjs';

const [, , entryPointArgument, ...restArguments] = process.argv;
if (!entryPointArgument) {
    process.stderr.write(
        'usage: extract.mjs <theme.css> [--json <path>] [--resolve-root <dir>] [--corpus-root <dir>]\n',
    );
    process.exit(2);
}
function flagValue(name) {
    const index = restArguments.indexOf(name);
    return index >= 0 ? restArguments[index + 1] : null;
}
const jsonOutputPath = flagValue('--json');

/*
 * The CSS entry point, the directory bare specifiers resolve from, and the tree scanned for a
 * corpus are three questions. For a repository stylesheet they collapse to one answer, which is why
 * they were one input and why the collapse stayed invisible.
 *
 * `--resolve-root` is the same split `generate_collapse` took in a0f8635, for the same reason:
 * a design system written to share no `@theme` with any repository has no `node_modules` above it,
 * so the file best able to show this tool describing bare Tailwind was the file it could not open.
 *
 * `--corpus-root` is separate because the derived one is four `dirname` calls up from the entry
 * point, which is right for a repository and arbitrary anywhere else. Pointed at a scratch file it
 * climbed into an unrelated tree and scanned 28,219 class occurrences that had nothing to do with
 * the design system under test, clearing the `>= 100` floor on somebody else's files. A scan that
 * cannot find a corpus should report zero, not borrow one.
 */
const resolveRootArgument = flagValue('--resolve-root');
const corpusRootArgument = flagValue('--corpus-root');

const { designSystem, tailwindVersion, entryPoint, recordedEntryPoint, resolveRoot, inferDataType } = await loadDesignSystem(
    entryPointArgument,
    resolveRootArgument,
);

if (typeof inferDataType !== 'function') {
    process.stderr.write(
        'extract.mjs: could not recover inferDataType from the installed Tailwind.\n' +
            'The descriptor model is a claim about that function, so without it there is nothing to test.\n' +
            'Refusing to fall back to a hand-rolled predicate, which would test the fallback instead.\n',
    );
    process.exit(3);
}

/*
 * The registry, and the roots.
 *
 * `getClassList()` is the design system's own enumeration, which is the population. Roots are read
 * back out of the parser rather than by splitting names on dashes, because a dash-splitter reads
 * `border-l` as root `border` with value `l`, which is the mistake that silently broke `border-x`
 * in an earlier table.
 *
 * A name can be both a static utility and a functional root — `flex` is `display: flex` and also
 * the root of `flex-4` — so both are recorded rather than one winning.
 */
const registryClassNames = [];
const modifierClassNames = [];
const functionalRoots = new Set();
const staticNames = new Set();

for (const entry of designSystem.getClassList?.() ?? []) {
    const name = Array.isArray(entry) ? entry[0] : typeof entry === 'string' ? entry : entry?.name;
    if (typeof name !== 'string') continue;
    registryClassNames.push(name);

    const asWritten = parseCandidate(designSystem, name);
    if (asWritten?.kind === 'static') staticNames.add(name);
    if (asWritten?.kind === 'functional' && asWritten.root) functionalRoots.add(asWritten.root);

    // `from-black/70` parses as root `from` and `from` never appears with a numeric probe, while
    // `flex` only reveals its functional root under one. Both routes are needed.
    const asFunctional = parseCandidate(designSystem, name + '-4');
    if (asFunctional?.kind === 'functional' && asFunctional.root) functionalRoots.add(asFunctional.root);

    /*
     * Modifiers are a second axis and are kept out of the registry count.
     *
     * `getClassList()` returns 37,643 entries, and many carry a modifier list: `bg-conic-0` alone
     * has eight. Expanding the cross product gives 639,612 names, which is a real space but it is
     * not the population the registry replay is measuring, and folding it in would make the headline
     * number unrecognisable against the 37,643 measured in the analysis. They are replayed
     * separately, sampled, so the modifier axis is still tested.
     */
    if (Array.isArray(entry) && entry[1] && typeof entry[1] === 'object') {
        for (const modifier of entry[1].modifiers ?? []) modifierClassNames.push(name + '/' + modifier);
    }
}

/*
 * The registry's own functional roots, which the class list does not advertise.
 *
 * `getClassList()` enumerates *classes*, and a root that advertises none is invisible to it however
 * the names are probed. Fourteen roots are in exactly that state here: `filter`, `backdrop-filter`,
 * `bg-size`, `bg-position`, `mask-position`, `font-features`, `flex-grow`, `flex-shrink`,
 * `max-w-screen`, `-col`, `-row`, `-hue-rotate` and `-backdrop-hue-rotate`. Every one is registered
 * functional and compiles — `filter-[var(--a)]` reads `[339]#1` — and none appears in the 37,643
 * entries the class list returns, so a root set derived from those entries structurally cannot
 * contain them and no probe suffix recovers them.
 *
 * `utilities.keys('functional')` is the registry's own enumeration and is the authority for which
 * roots exist. It reports 341 where the class list yields 327, and the 14 it adds are exactly the
 * ones above.
 *
 * They were not merely absent, they were absent *invisibly*: a class naming one of them parses to
 * the root correctly and then finds no descriptor row, so `Lookup` declines, which reads in every
 * summary as the model's own declared boundary rather than as a table gap. Only `max-w-screen`
 * surfaced, on www-connected-app, and only because that registry advertises `max-w-screen-sm` while
 * ahra's does not. The other thirteen were latent, waiting on a repository writing a class that
 * happened to reach one.
 *
 * `max-w-screen` is also the case that shows why both sources are needed rather than one replacing
 * the other. It is registered as a static utility *and* as a functional root; the class list route
 * records the static, the registry route records the functional, and dropping either loses a real
 * reading.
 */
for (const root of designSystem.utilities?.keys?.('functional') ?? []) {
    if (typeof root === 'string') functionalRoots.add(root);
}

const roots = Array.from(functionalRoots).sort();

/*
 * The bare values each root is actually written with, taken from the registry.
 *
 * Some roots define literal keywords of their own that are in no theme namespace and infer as
 * nothing: `transition-none` sets `transition-property: none` and reads `[350]#1` while every other
 * `transition-*` reads `[350,353,354]#3`; `ease-initial` reads `[]#1` against `ease-linear`'s
 * `[354]#2`. There is nothing to look them up in, so they are read off the registry, which is the
 * enumeration of exactly the names the root accepts.
 *
 * This is per-root data rather than a rule, which is the honest shape: the keywords are literals in
 * the utility's own handler, and a table is what a literal becomes.
 */
const registryValuesByRoot = new Map();
for (const className of registryClassNames) {
    const candidate = parseCandidate(designSystem, className);
    if (candidate?.kind !== 'functional' || !candidate.root) continue;
    if (candidate.modifier != null) continue;
    const value = candidate.value;
    if (!value || value.kind === 'arbitrary') continue;
    const list = registryValuesByRoot.get(candidate.root) ?? new Set();
    list.add(String(value.value ?? ''));
    registryValuesByRoot.set(candidate.root, list);
}

/*
 * The theme's namespaces, and a few keys from each.
 *
 * Namespaces are derived from the theme's own keys rather than hardcoded, because half of them are
 * this repo's (`--color-content-*`, `--percentage-*`) and a hardcoded list would be right for one
 * repo. A key is `--<namespace>-<name>`, and the namespace boundary is ambiguous from the string
 * alone — `--color-red-500` could split three ways — so each candidate prefix is confirmed against
 * `keysInNamespaces`, the design system's own membership test.
 */
const themeKeysByNamespace = new Map();
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
        themeKeysByNamespace.set(prefix, keys);
    }
}

/*
 * Longest namespace first.
 *
 * `--drop-shadow` and `--shadow` both match `drop-shadow-lg`'s key space, and the root resolves the
 * more specific one. Probing in longest-first order means the specific namespace claims its reading
 * before the general one can, which is the same precedence the engine applies.
 */
/*
 * Tailwind's global property order, read out of the installed bundle.
 *
 * Only arbitrary properties need it here, and only to turn a property name into an index. It ships
 * as a plain string array, so it is recovered exactly. The same extraction is in
 * `internal/lint/rules/tailwind/tools/generate_collapse/enumerate.mjs`, whose comment records why the obvious alternative
 * fails: deriving the order by asking the engine to sort one utility per property only reaches
 * properties some utility declares alone, which found 254 of 359 and shifted every index after a
 * missing one.
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
    if (orderMatch !== null) {
        for (const property of JSON.parse(orderMatch[0])) propertyOrder.push(property);
    }
}

const themeNamespaces = Array.from(themeKeysByNamespace.keys()).sort((left, right) => right.length - left.length || (left < right ? -1 : 1));

/*
 * The keys that belong to exactly one namespace.
 *
 * `--color` and `--color-red` both hold `red-500`, and `--font-weight` and `--color` both hold
 * `black`. A shared key cannot tell you which namespace a root is reading, so discovery uses only
 * the unambiguous ones. Prediction still consults every namespace, because the question there is
 * membership rather than attribution.
 */
const uniqueThemeKeysByNamespace = new Map();
{
    const namespacesByKey = new Map();
    for (const [namespace, keys] of themeKeysByNamespace) {
        for (const key of keys) namespacesByKey.set(key, (namespacesByKey.get(key) ?? 0) + 1);
    }
    for (const [namespace, keys] of themeKeysByNamespace) {
        uniqueThemeKeysByNamespace.set(namespace, keys.filter((key) => namespacesByKey.get(key) === 1));
    }
}

/*
 * The inferred type of a probe, as the engine would see it.
 *
 * Two paths, and conflating them is the trap. A typed hint (`[color:red]`) names the type outright
 * and the parser records it on the candidate as `value.dataType`, bypassing inference entirely. A
 * bare arbitrary value is inferred, and inference is *relative to a type list* — `inferDataType`
 * walks the list in order and returns the first match, so the same value types differently under
 * different roots. That is why the descriptor carries a type list rather than a single type.
 *
 * Returns the string `'@bare'` for non-arbitrary values, which route through theme lookup instead of
 * inference and are a different axis, and null when inference declines (every `var(...)`, and
 * anything matching nothing).
 */
function inferredTypeOf(className, typeList) {
    const candidate = parseCandidate(designSystem, className);
    if (!candidate) return undefined;
    const value = candidate.value;
    if (!value) return '@empty';
    if (value.kind !== 'arbitrary') return '@bare:' + String(value.value ?? '');
    if (value.dataType) return value.dataType;
    const inferred = inferDataType(String(value.value ?? ''), typeList);
    return inferred ?? null;
}

/*
 * Recover a root's type-list order by observation.
 *
 * A root's type list is what the source writes as `inferDataType(value, ['color','length'])`, and
 * those lists live inside 21 call sites in a 6,827-line file that ships minified. Reading them would
 * mean parsing that; observing them tests the engine instead of the transcription.
 *
 * Membership is decided in `discoverDescriptor` — a type is in the list when it changes a reading on
 * some axis. This function decides the order, which matters because `inferDataType` returns the
 * FIRST match.
 *
 * For every value that more than one of the root's types accepts, the engine's reading names the
 * winner, and the winner is earlier in the list. Those pairwise facts are a partial order; a stable
 * topological sort turns them into the list. Types never disambiguated by any probe keep their
 * declaration-order position, which is harmless precisely because nothing observable distinguishes
 * them.
 */
function orderTypeList(root, types, readingByType, fallbackReading) {
    if (types.length < 2) return types;

    // Probe values that several predicates accept. Drawn from the shape corpus rather than a
    // separate list, so a shape added for the sweep also sharpens the ordering.
    const ambiguousProbes = new Set();
    for (const shape of Shapes) {
        if (!shape.syntax.startsWith('[') || shape.syntax.includes(':')) continue;
        const inner = shape.syntax.slice(1, -1);
        let matches = 0;
        for (const typeName of types) {
            if (inferDataType(inner, [typeName]) !== null) matches++;
            if (matches > 1) break;
        }
        if (matches > 1) ambiguousProbes.add(inner);
    }

    // `earlier.get(a)` holds every type observed to lose to `a`.
    const earlier = new Map(types.map((typeName) => [typeName, new Set()]));
    for (const probe of ambiguousProbes) {
        const actual = readingOf(designSystem, root + '-[' + probe + ']');
        if (actual === null) continue;

        const accepting = types.filter((typeName) => inferDataType(probe, [typeName]) !== null);
        // The winners: types whose reading matches what the engine produced. More than one can match
        // when two types share a reading, and then the probe says nothing about their relative order.
        /*
         * Winners and losers, where either side can hold several types.
         *
         * A type absent from `readingByType` reads like the fallback, so it wins when the engine
         * produced the fallback. Several types can share a reading and all win together — `-3` is
         * accepted by `family-name`, `number` and `line-width` for `font`, and the last two both read
         * the fallback. A first version required exactly one winner and threw the constraint away,
         * which left `family-name` ahead of `number` and predicted `font-[-3]` as a font family.
         *
         * What the probe proves is that every winner precedes every loser. It says nothing about the
         * order within either group, and nothing is recorded about that.
         */
        const winners = accepting.filter((typeName) => sameReading(readingByType.get(typeName) ?? fallbackReading, actual));
        if (winners.length === 0 || winners.length === accepting.length) continue;
        for (const winner of winners) {
            for (const loser of accepting) {
                if (!winners.includes(loser)) earlier.get(winner).add(loser);
            }
        }
    }

    // Stable topological sort: repeatedly take the first type nothing unplaced beats.
    const ordered = [];
    const remaining = types.slice();
    while (remaining.length > 0) {
        let index = remaining.findIndex((typeName) => !remaining.some((other) => other !== typeName && earlier.get(other).has(typeName)));
        // A cycle means two probes disagree about precedence, which the descriptor model cannot
        // express. Taking the first entry keeps the sort total; the sweep then reports the failures
        // rather than hiding them behind an arbitrary order.
        if (index === -1) index = 0;
        ordered.push(remaining[index]);
        remaining.splice(index, 1);
    }
    return ordered;
}

function discoverDescriptor(root) {
    /*
     * The fallback: what the root reads when inference declines.
     *
     * `var()` is the canonical case, since `inferDataType` refuses to inspect it by an explicit early
     * return. Some roots reject it outright — `shadow-` is a repo `@utility` whose `--value()` list
     * does not admit an arbitrary value, so `shadow--[var(--a)]` compiles to nothing — and for those
     * the fallback is read from a bare value the root does accept. Without it the descriptor had no
     * fallback at all and predicted null for four probes it should have answered.
     */
    let fallbackReading = readingOf(designSystem, root + '-[var(--tw-probe)]');
    if (fallbackReading === null) {
        for (const probeValue of ['[anything]', '-1', '1', '4', 'none', 'full']) {
            const reading = readingOf(designSystem, root + '-' + probeValue);
            if (reading === null) continue;
            fallbackReading = reading;
            break;
        }
    }

    /*
     * The types the root actually discriminates on, which is not the types it accepts.
     *
     * Every root accepts every hint: `drop-shadow-[vector:var(--a)]` compiles, because an
     * unrecognised hint is not rejected, it falls through to the root's default handling. So a first
     * version recorded a reading for all 17 types on every root and produced type lists like
     * `drop-shadow: [color, length, percentage, ratio, number, ..., vector]` — which is not the
     * `inferDataType(value, ['color'])` written in the source.
     *
     * That over-broad list is not merely untidy, it is wrong, and it fails loudly on the bare path.
     * `family-name` is a permissive predicate — it accepts any comma-separated run not starting with
     * a digit — so under a 17-type list `inferDataType('amber-500', types)` returns `family-name`
     * and `drop-shadow-amber-500` is predicted as a font. Registry agreement fell from 37,446 to
     * 25,712 the moment inference was let near the bare path with these lists.
     *
     * A type is in the list only when its reading differs from the fallback. A type that reads the
     * same as the fallback carries no information: whether the root "supports" it is unobservable
     * through `{order, count}`, which is the only thing being modelled.
     */
    const readingByType = new Map();
    const acceptedTypes = [];
    for (const typeName of DataTypeNames) {
        const reading = readingOf(designSystem, root + '-[' + typeName + ':var(--tw-probe)]');
        if (reading === null) continue;
        acceptedTypes.push(typeName);
        if (sameReading(reading, fallbackReading)) continue;
        readingByType.set(typeName, reading);
    }

    /*
     * A type that reads like the fallback still belongs in the list.
     *
     * It carries no reading of its own, but it carries a *position*: a value it claims is a value
     * the types after it never see. `font`'s real list is `['number','generic-name','family-name']`
     * and `number`'s reading is the fallback, so dropping it let `family-name` claim `-3` and
     * `font-[-3]` was predicted as a font family instead of a weight. Keeping it in the order and
     * out of `readingByType` gets both right: it shadows, and it resolves to the fallback.
     *
     * Only types that overlap a type with a distinct reading are worth keeping — the rest add
     * ordering constraints nothing observes, and each one costs a row in the generated table.
     */
    const shadowingTypes = acceptedTypes.filter((typeName) => {
        if (readingByType.has(typeName)) return false;
        return Array.from(readingByType.keys()).some((other) => Shapes.some((shape) => {
            if (!shape.syntax.startsWith('[') || shape.syntax.includes(':')) return false;
            const inner = shape.syntax.slice(1, -1);
            return inferDataType(inner, [typeName]) !== null && inferDataType(inner, [other]) !== null;
        }));
    });

    /*
     * The bare axis: theme values rather than arbitrary ones.
     *
     * A bare value is not inferred, it is resolved. `drop-shadow-amber-500` and `drop-shadow-lg`
     * carry the same syntax and read differently, and no amount of `inferDataType` explains it,
     * because `amber-500` is not a value at all until the theme turns it into one. What discriminates
     * them is the *namespace* it resolves in: `--color-amber-500` exists, `--drop-shadow-amber-500`
     * does not, and the root tries its namespaces in a fixed order.
     *
     * A first version probed one bare value per root and used its reading for all of them, which
     * agreed on 15,251 of 37,643 registry classes and disagreed on 22,392 — every color-valued class
     * of every root that also takes a non-color theme value. The single-probe reading was not wrong
     * about the root, it was answering a question the root does not have one answer to.
     *
     * So the bare axis is partitioned the same way the arbitrary one is: probe the root with a
     * representative key from each theme namespace and keep the namespaces whose readings differ.
     * `keysInNamespaces` is the design system's own membership test, so this reads the theme rather
     * than guessing at it from key prefixes.
     */
    const readingByNamespace = new Map();
    for (const namespace of themeNamespaces) {
        const keys = uniqueThemeKeysByNamespace.get(namespace);
        if (!keys || keys.length === 0) continue;
        /*
         * Several keys, because one key can be absent from the root's own accepted set while another
         * in the same namespace is present, and only keys unique to this namespace, because a shared
         * key attributes another namespace's reading to this one.
         *
         * `bold` is in `--font-weight`; `black` is in `--font-weight` and in `--color`. Probing
         * `font-black` and recording the answer under `--color` gave `font` a `--color` entry with
         * font-weight's reading, and since `--color` is 7 characters and `--font-weight` is 13, the
         * longest-first lookup still resolved `bold` correctly but `font-sans` did not — `--font` is
         * shorter than the spurious entries above it. Eight registry classes turned on this.
         */
        for (const key of keys) {
            const reading = readingOf(designSystem, root + '-' + key);
            if (reading === null) continue;
            readingByNamespace.set(namespace, reading);
            break;
        }
    }

    /*
     * The non-namespaced bare bucket.
     *
     * Spacing roots take `p-4`, and `4` is in no namespace: it multiplies `--spacing`. Fractions
     * (`w-1/2`), keywords the root itself defines (`w-full`), and bare numbers all land here. Probed
     * with a spread of shapes rather than one, since a root accepts only some of them.
     */
    for (const bareValue of ['4', '1', '0', '0.5', '2.5', 'full', '1/2', '3', '96']) {
        const reading = readingOf(designSystem, root + '-' + bareValue);
        if (reading === null) continue;
        readingByNamespace.set('@none', reading);
        break;
    }

    /*
     * The universal color keywords, which are in no namespace and infer as nothing.
     *
     * `border-current` reads as a color and `border-4` does not, but `current` is absent from
     * `--color` (verified: `keysInNamespaces(['--color'])` holds 558 keys and `current` is not one)
     * and `inferDataType('current', ['color'])` returns null. They are Tailwind's own built-ins,
     * resolved by the color path before the theme is consulted at all.
     *
     * They cost their own bucket because folding them into `@none` mispredicted every one of them:
     * 197 registry classes, which is what remained after the namespace axis landed. `transparent` is
     * listed even though it does infer as a color, so the bucket is the whole keyword set rather
     * than the subset inference happens to miss.
     */
    for (const keyword of ['current', 'inherit', 'transparent']) {
        const reading = readingOf(designSystem, root + '-' + keyword);
        if (reading === null) continue;
        readingByNamespace.set('@colorKeyword', reading);
        break;
    }

    /*
     * Literal keywords: registry values that no other bucket explains.
     *
     * Recorded only when the reading differs from what the rest of the descriptor would predict, so
     * the table carries an entry per genuine exception rather than per registry value. On both repos
     * probed this is two entries in total, `transition-none` and `ease-initial`.
     */
    const readingByLiteral = new Map();
    for (const bareValue of registryValuesByRoot.get(root) ?? []) {
        if (ColorKeywords.has(bareValue)) continue;
        if (themeNamespaces.some((namespace) => readingByNamespace.has(namespace) && themeKeysByNamespace.get(namespace).includes(bareValue))) continue;
        /*
         * Tested against the types this root discriminates on, not against all of them.
         *
         * `family-name` accepts any comma-separated run not starting with a digit, so a guard over
         * the whole type table types `none` and `initial` as font families and rejects every literal
         * before it can be recorded. The question is only whether some bucket the descriptor already
         * has explains this value.
         */
        const inferredHere = inferDataType(bareValue, [...readingByType.keys()]);
        if (inferredHere !== null && readingByType.has(inferredHere)) continue;
        const reading = readingOf(designSystem, root + '-' + bareValue);
        if (reading === null) continue;
        if (sameReading(reading, readingByNamespace.get('@none') ?? fallbackReading)) continue;
        readingByLiteral.set(bareValue, reading);
    }

    /*
     * The empty reading: the root written with no value at all, which some roots accept.
     *
     * It takes the modifier axis too. `shadow` alone reads `[315,316]#2` and `shadow/5` reads
     * `[315,316]#3`, because the alpha is a declaration whether or not there is a value to apply it
     * to. Two of the 40,110 sampled modifier classes turned on this.
     */
    const emptyReading = readingOf(designSystem, root);
    const emptyReadingWithModifier = readingOf(designSystem, root + '/50');
    const emptyReadingWithThemedModifier = readingOf(designSystem, root + '/none');

    /*
     * The modifier axis, which the descriptor model as originally stated does not have.
     *
     * `bg-red-500/50` and `bg-red-500` share a root and a data type, so the model predicts one
     * reading for both, and for `bg` it is right. For `text` and `shadow` it is not:
     *
     *   text-[3px]      [286]#1        text-[3px]/50      [286,287]#2
     *   shadow-[3px]    [315,316]#2    shadow-[3px]/50    [315,316]#3
     *
     * A modifier on `text` emits `line-height` alongside `font-size`; on `shadow` it emits an alpha
     * declaration. Both are extra declarations, so both move the reading, and no amount of type
     * inference sees them because the modifier is not part of the value.
     *
     * This is a real addition to Phase 3's descriptor and it is a small one: the delta is a property
     * of the root, not of the value. Recorded as a whole reading rather than a delta so the lookup
     * stays a map read.
     */
    /*
     * Probed over every accepted type, not only the discriminating ones.
     *
     * Whether a modifier moves the reading is itself type-dependent: `text-shadow-[red]/50` reads
     * the same as `text-shadow-[red]`, while `text-shadow-[3px]/50` gains a declaration. `color` is
     * not in `text-shadow`'s `readingByType` — unmodified, it reads like the fallback — so a
     * modifier map keyed only on discriminating types had no entry for it and fell through to the
     * modified fallback, which is the `[3px]` answer. Six probes turned on that.
     *
     * So the modifier axis is keyed on the full accepted set, and a type can discriminate under a
     * modifier while carrying no information without one.
     */
    const readingByTypeWithModifier = new Map();
    for (const typeName of acceptedTypes) {
        const reading = readingOf(designSystem, root + '-[' + typeName + ':var(--tw-probe)]/50');
        if (reading !== null) readingByTypeWithModifier.set(typeName, reading);
    }
    const fallbackReadingWithModifier = readingOf(designSystem, root + '-[var(--tw-probe)]/50');

    /*
     * A third bucket, for a modifier whose own value resolves through the theme.
     *
     * `/50` is an alpha and every alpha behaves alike: `/25`, `/[0.5]` and `/[var(--a)]` all give the
     * same reading, which is what makes the axis one bit. `/none` is not an alpha. It is a
     * `--leading` key, so `text-[3px]/none` emits `line-height: normal` and reads like any other
     * modified `text`, while `shadow-[3px]/none` finds nothing to resolve and reads like an
     * unmodified `shadow`.
     *
     * So `none` is not a special case bolted onto the model, it is the same theme-resolution the
     * value axis already does, applied on the modifier. It gets a bucket rather than a branch, and
     * the axis is `{absent, alpha, themed}` instead of `{absent, present}`.
     */
    const readingByTypeWithThemedModifier = new Map();
    for (const typeName of acceptedTypes) {
        const reading = readingOf(designSystem, root + '-[' + typeName + ':var(--tw-probe)]/none');
        if (reading !== null) readingByTypeWithThemedModifier.set(typeName, reading);
    }
    const fallbackReadingWithThemedModifier = readingOf(designSystem, root + '-[var(--tw-probe)]/none');
    const readingByNamespaceWithModifier = new Map();
    const readingByNamespaceWithThemedModifier = new Map();
    for (const [namespace] of readingByNamespace) {
        const keys = namespace === '@none'
            ? ['4', '1', '0', '0.5', '2.5', 'full', '1/2', '3', '96']
            : namespace === '@colorKeyword'
              ? ['current', 'inherit', 'transparent']
              : (uniqueThemeKeysByNamespace.get(namespace) ?? []);
        for (const key of keys) {
            const reading = readingOf(designSystem, root + '-' + key + '/50');
            if (reading === null) continue;
            readingByNamespaceWithModifier.set(namespace, reading);
            break;
        }
        for (const key of keys) {
            const reading = readingOf(designSystem, root + '-' + key + '/none');
            if (reading === null) continue;
            readingByNamespaceWithThemedModifier.set(namespace, reading);
            break;
        }
    }

    /*
     * The type list, in the order the root applies it.
     *
     * `inferDataType` returns the FIRST type in the list that matches, so the list is ordered and
     * the order is per-root. `bg`'s is `['image','percentage','position','bg-size','length','url']`,
     * and `3px` satisfies both `position` and `length`: under Tailwind's order it is a position and
     * reads `[254]`, under this file's declaration order it is a length and reads `[251]`. Fifty-
     * eight `bg` probes turned on that single inversion, which is why `bg` was the canary — it is
     * the root with the most types whose predicates overlap.
     *
     * The order is recovered rather than transcribed, by topological sort from observation: for each
     * pair of types whose predicates both accept some probe value, the type whose reading the engine
     * actually produced comes first. Transcribing it would mean reading 21 call sites out of a
     * minified 6,827-line file; observing it tests the engine instead of the transcription.
     */
    /*
     * A type discriminating only under a modifier still belongs in the list.
     *
     * `text-shadow-[red]` and `text-shadow-[3px]` read alike, so `color` carries no information and
     * was dropped; but `text-shadow-[red]/50` and `text-shadow-[3px]/50` do not, so under a modifier
     * `color` is exactly what tells them apart. The list is one list across all three modifier
     * states, and `inferDataType` only ever consults that one list, so a type earning its place on
     * any axis has to be on it.
     */
    const modifierDiscriminatingTypes = acceptedTypes.filter((typeName) => {
        if (readingByType.has(typeName) || shadowingTypes.includes(typeName)) return false;
        const modified = readingByTypeWithModifier.get(typeName);
        const themed = readingByTypeWithThemedModifier.get(typeName);
        return (modified !== undefined && !sameReading(modified, fallbackReadingWithModifier))
            || (themed !== undefined && !sameReading(themed, fallbackReadingWithThemedModifier));
    });

    const typeList = orderTypeList(root, [...readingByType.keys(), ...shadowingTypes, ...modifierDiscriminatingTypes], readingByType, fallbackReading);

    return {
        root,
        typeList,
        readingByType,
        fallbackReading,
        readingByNamespace,
        readingByLiteral,
        emptyReading,
        emptyReadingWithModifier,
        emptyReadingWithThemedModifier,
        readingByTypeWithModifier,
        fallbackReadingWithModifier,
        readingByNamespaceWithModifier,
        readingByTypeWithThemedModifier,
        fallbackReadingWithThemedModifier,
        readingByNamespaceWithThemedModifier,
    };
}

/*
 * Predict a class's reading from its descriptor.
 *
 * This is the Go lookup written in JavaScript, and it is deliberately the whole model: parse, infer
 * a type against the root's own type list, take that type's reading, and otherwise take the
 * fallback. If this function predicts every reading the engine produces, the Go port is a
 * transcription. Every place it does not is a place the port needs more than a table.
 */
/*
 * Arbitrary properties: a fourth candidate kind, with no root and no descriptor.
 *
 * `[font:inherit]` is not a utility at all, it is a bare CSS declaration written as a class, and the
 * parser reports `kind: 'arbitrary'` for it. Its reading is one declaration whose property is the
 * text before the colon, so there is nothing to look up and nothing to generate: the Go side reads
 * the property name out of the candidate and indexes `PropertyOrder` directly.
 *
 * It gets a branch rather than a table row, and it is here because the corpus contains one and a
 * harness that could not predict it would be reporting a model failure for something the model was
 * never asked about.
 */
function arbitraryPropertyReading(candidate) {
    const property = String(candidate.property ?? '');
    const index = propertyOrder.indexOf(property);
    return { order: index === -1 ? [] : [index], count: 1 };
}

/*
 * Tailwind's built-in color keywords: in no theme namespace, inferring as nothing (except
 * `transparent`), and read as colors regardless. Enumerated rather than derived because there is
 * nothing to derive them from — they are literals in `utilities.ts`'s color path.
 */
const ColorKeywords = new Set(['current', 'inherit', 'transparent']);

/*
 * Modifier values that resolve through the theme rather than as an alpha.
 *
 * `none` is the only one in this Tailwind: it is a `--leading` key, so it means something to `text`
 * and nothing to `shadow`. Enumerated rather than derived because deriving it would mean asking
 * which namespaces a modifier consults, which is per-root, while the set of non-alpha modifier
 * values is not.
 */
const ThemedModifierValues = new Set(['none']);

function predictReading(descriptor, className) {
    const inferred = inferredTypeOf(className, descriptor.typeList);
    if (inferred === undefined) return undefined;
    /*
     * A modifier selects a parallel set of readings rather than adjusting one.
     *
     * The modifier's *value* is not consulted: `/50`, `/[0.5]` and `/[var(--a)]` all give the same
     * reading, because what a modifier adds is a declaration and the declaration exists whatever the
     * alpha is. So the axis is one bit, present or absent, which is what makes it cheap enough to
     * add to every descriptor.
     */
    const candidate = parseCandidate(designSystem, className);
    const modifier = candidate?.modifier ?? null;
    // `none` is a theme key rather than an alpha; anything else is an alpha and they all behave alike.
    const modifierAxis = modifier === null ? 'absent' : modifier.kind === 'named' && ThemedModifierValues.has(modifier.value) ? 'themed' : 'alpha';

    const readingByType = modifierAxis === 'absent'
        ? descriptor.readingByType
        : modifierAxis === 'themed'
          ? descriptor.readingByTypeWithThemedModifier
          : descriptor.readingByTypeWithModifier;
    const readingByNamespace = modifierAxis === 'absent'
        ? descriptor.readingByNamespace
        : modifierAxis === 'themed'
          ? descriptor.readingByNamespaceWithThemedModifier
          : descriptor.readingByNamespaceWithModifier;
    const fallbackReading = modifierAxis === 'absent'
        ? descriptor.fallbackReading
        : modifierAxis === 'themed'
          ? descriptor.fallbackReadingWithThemedModifier
          : descriptor.fallbackReadingWithModifier;

    if (inferred === '@empty') {
        return modifierAxis === 'absent'
            ? descriptor.emptyReading
            : modifierAxis === 'themed'
              ? descriptor.emptyReadingWithThemedModifier
              : descriptor.emptyReadingWithModifier;
    }

    if (inferred !== null && inferred.startsWith('@bare:')) {
        /*
         * The bare path, in the precedence the engine applies.
         *
         * It is four steps, and each was added because the one before it left a measurable
         * population mispredicted:
         *
         *   1. The color keywords. `current` and `inherit` are in no namespace and infer as nothing,
         *      yet read as colors, so they have to be tested before anything can claim them.
         *   2. The theme namespaces, longest first, since `--drop-shadow` must win over `--shadow`.
         *   3. Inference. Bare values are typed too — `via-25%` reads as a percentage. The single
         *      biggest mistake available here is assuming bare means "theme lookup only".
         *   4. Everything else: bare numbers, fractions, root-defined keywords like `full`.
         *
         * Theme before inference, and the order is not arbitrary. `bold` is a `--font-weight` key
         * and it also satisfies `family-name`, which is in `font`'s type list; inferring first
         * predicted `font-bold` as a font family. Eight registry classes distinguish the two
         * orderings, and the theme is what the engine consults first.
         */
        const bareValue = inferred.slice('@bare:'.length);

        if (modifierAxis === 'absent' && descriptor.readingByLiteral.has(bareValue)) {
            return descriptor.readingByLiteral.get(bareValue);
        }

        if (readingByNamespace.has('@colorKeyword') && ColorKeywords.has(bareValue)) {
            return readingByNamespace.get('@colorKeyword');
        }

        for (const namespace of themeNamespaces) {
            if (!readingByNamespace.has(namespace)) continue;
            if (!themeKeysByNamespace.get(namespace).includes(bareValue)) continue;
            return readingByNamespace.get(namespace);
        }

        const inferredBareType = inferDataType(bareValue, descriptor.typeList);
        if (inferredBareType !== null && readingByType.has(inferredBareType)) {
            return readingByType.get(inferredBareType);
        }

        return readingByNamespace.get('@none') ?? fallbackReading;
    }
    if (inferred !== null && readingByType.has(inferred)) {
        return readingByType.get(inferred);
    }
    return fallbackReading;
}

function sameReading(left, right) {
    if (left === null || right === null || left === undefined || right === undefined) {
        return left === right;
    }
    return left.count === right.count && left.order.length === right.order.length && left.order.every((value, index) => value === right.order[index]);
}

// ---------------------------------------------------------------------------------------------
// Pass 1 — extract

const descriptors = new Map();
for (const root of roots) descriptors.set(root, discoverDescriptor(root));

// Static utilities are not a model at all: they take no value, so their reading is a constant. They
// are extracted so the table Phase 3 consumes is complete, and they are replayed like everything
// else so a constant that is not constant would show up.
const staticReadings = new Map();
for (const name of staticNames) {
    const reading = readingOf(designSystem, name);
    if (reading !== null) staticReadings.set(name, reading);
}

// ---------------------------------------------------------------------------------------------
// Pass 2 — replay against the registry

const registryResult = {
    measured: 0,
    agreed: 0,
    nullReading: 0,
    unparsed: 0,
    disagreed: [],
};

for (const className of registryClassNames) {
    const candidate = parseCandidate(designSystem, className);
    if (!candidate) {
        registryResult.unparsed++;
        continue;
    }

    const actual = readingOf(designSystem, className);
    if (actual === null) {
        // No reading at all. Not an agreement, and counted on its own line so a run cannot look
        // clean by measuring nothing.
        registryResult.nullReading++;
        continue;
    }

    let predicted;
    if (candidate.kind === 'arbitrary') {
        predicted = arbitraryPropertyReading(candidate);
    }
    else if (candidate.kind === 'static') {
        predicted = staticReadings.get(className) ?? null;
    }
    else {
        const descriptor = descriptors.get(candidate.root);
        predicted = descriptor ? predictReading(descriptor, className) : undefined;
    }

    registryResult.measured++;
    if (sameReading(predicted, actual)) {
        registryResult.agreed++;
        continue;
    }
    registryResult.disagreed.push({
        className,
        root: candidate.root ?? null,
        kind: candidate.kind,
        predicted: readingKey(predicted ?? null),
        actual: readingKey(actual),
    });
}

/*
 * Pass 2b — the modifier axis.
 *
 * The descriptor model says nothing about modifiers: it maps a root and a data type to a reading,
 * and `bg-red-500/50` has the same root and the same type as `bg-red-500`. So either modifiers do
 * not move the reading, in which case the model is complete without a modifier axis, or they do, in
 * which case Phase 3 needs one. Sampled rather than exhaustive because the full cross product is
 * 601,969 names for one bit of information.
 */
const modifierResult = { measured: 0, agreed: 0, nullReading: 0, disagreed: [] };
{
    const stride = Math.max(1, Math.floor(modifierClassNames.length / 40000));
    for (let index = 0; index < modifierClassNames.length; index += stride) {
        const className = modifierClassNames[index];
        const candidate = parseCandidate(designSystem, className);
        if (!candidate) continue;

        const actual = readingOf(designSystem, className);
        if (actual === null) {
            modifierResult.nullReading++;
            continue;
        }

        const descriptor = candidate.kind === 'functional' ? descriptors.get(candidate.root) : null;
        const predicted = candidate.kind === 'static' ? (staticReadings.get(className.split('/')[0]) ?? null) : descriptor ? predictReading(descriptor, className) : undefined;

        modifierResult.measured++;
        if (sameReading(predicted, actual)) {
            modifierResult.agreed++;
            continue;
        }
        if (modifierResult.disagreed.length < 5000) {
            modifierResult.disagreed.push({
                className,
                root: candidate.root ?? null,
                predicted: readingKey(predicted ?? null),
                actual: readingKey(actual),
            });
        }
    }
}

// ---------------------------------------------------------------------------------------------
// Pass 3 — sweep the arbitrary value space

const sweepResult = {
    probed: 0,
    measured: 0,
    skippedUnregistered: 0,
    agreed: 0,
    nullReading: 0,
    disagreed: [],
    readingsPerRoot: new Map(),
};

for (const root of roots) {
    const descriptor = descriptors.get(root);
    const distinctReadings = new Set();

    for (const shape of Shapes) {
        const className = probeClassName(root, shape);
        sweepResult.probed++;

        const actual = readingOf(designSystem, className);
        if (actual === null) {
            sweepResult.nullReading++;
            continue;
        }
        distinctReadings.add(readingKey(actual));

        /*
         * A probe on a functional root can land on a static utility.
         *
         * `w` is a functional root and `w-px` is not `w` with the value `px` — it is a static
         * utility whose whole name is `w-px`, and the parser says so. Routing it through the
         * descriptor asked the wrong table and mispredicted 349 probes across `w`, `h`, `min-w`,
         * `block` and friends, all of them the same handful of names (`px`, `full`, `auto`, `none`,
         * `screen`, `fit`).
         *
         * The descriptor model is not implicated. The probe simply was not a probe of this root.
         */
        /*
         * The probe's real root, read from the parser rather than assumed to be the one being swept.
         *
         * A probe on a functional root does not necessarily land on that root. `w-px` is a static
         * utility whose whole name is `w-px`; `border-x` is its own functional root, not `border`
         * with the value `x`; `content--1` parses as root `content-` with value `1`. Attributing any
         * of those to the root under sweep asks the wrong table and reports a model failure where
         * there is only a misrouted probe — 349 of them, all the same handful of names.
         *
         * A probe that lands somewhere else is still measured, against the descriptor that actually
         * owns it. Skipping it instead would quietly shrink the population.
         */
        const probeCandidate = parseCandidate(designSystem, className);
        /*
         * A static the registry never enumerated is not a claim about anything.
         *
         * `max-w-screen` and `order-none` parse as static utilities, but neither is in
         * `getClassList()`, so no extractor could have a row for them and predicting null is correct
         * rather than wrong. They exist because the sweep composes `root + shape` and two shapes
         * happen to spell a name Tailwind's parser recognises.
         */
        if (probeCandidate?.kind === 'static' && !staticReadings.has(className)) {
            sweepResult.skippedUnregistered++;
            continue;
        }
        const probeDescriptor = probeCandidate?.kind === 'functional' && probeCandidate.root !== root
            ? descriptors.get(probeCandidate.root)
            : descriptor;
        const predicted = probeCandidate?.kind === 'static'
            ? (staticReadings.get(className) ?? null)
            : probeDescriptor
              ? predictReading(probeDescriptor, className)
              : undefined;
        sweepResult.measured++;
        if (sameReading(predicted, actual)) {
            sweepResult.agreed++;
            continue;
        }
        if (sweepResult.disagreed.length < 20000) {
            sweepResult.disagreed.push({
                className,
                root,
                shape: shape.syntax,
                group: shape.group,
                inferred: String(inferredTypeOf(className, descriptor.typeList)),
                predicted: readingKey(predicted ?? null),
                actual: readingKey(actual),
            });
        }
    }

    sweepResult.readingsPerRoot.set(root, distinctReadings);
}

/*
 * Pass 2c — the corpus, for the null count.
 *
 * The registry is the population and it holds only classes the engine can compile, so its null count
 * is zero and always will be. The documented figure — 10 of 1,208 distinct classes returning null —
 * came from the *corpus*, the classes this codebase actually writes, which is a different population
 * containing things the registry never will: `group` and `peer` generate no CSS of their own,
 * `prose-lg` and `gmail_quote` are not Tailwind classes at all, `text-dark-4/70` names a theme key
 * that does not exist.
 *
 * Reporting only the registry's zero would be true and misleading. The corpus is measured separately
 * so the number that matters is the number that appears: a class returning no reading is not an
 * agreement, and the count of them is reported rather than folded into either column.
 *
 * Extracted with a regex over `className="..."` rather than a parser. That under-counts — it misses
 * template literals, `clsx` calls and conditional classes — and it is sufficient, because the
 * question here is whether the null population is non-empty and named, not how large it is.
 */
const corpusResult = { scanned: 0, distinct: 0, measured: 0, agreed: 0, nullReading: 0, disagreed: [], nullExamples: [] };
{
    const corpusRoot = corpusRootArgument
        ? NodePath.resolve(corpusRootArgument)
        : NodePath.dirname(NodePath.dirname(NodePath.dirname(NodePath.dirname(entryPoint))));
    const distinctClasses = new Set();

    function scanDirectory(directory, depth) {
        if (depth > 8) return;
        let entries;
        try {
            entries = NodeFileSystem.readdirSync(directory, { withFileTypes: true });
        }
        catch {
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
            }
            catch {
                continue;
            }
            for (const match of contents.matchAll(/className="([^"]*)"/g)) {
                for (const className of match[1].split(/\s+/)) {
                    if (className.length === 0) continue;
                    corpusResult.scanned++;
                    distinctClasses.add(className);
                }
            }
        }
    }
    scanDirectory(corpusRoot, 0);
    corpusResult.distinct = distinctClasses.size;

    for (const className of distinctClasses) {
        const actual = readingOf(designSystem, className);
        if (actual === null) {
            corpusResult.nullReading++;
            if (corpusResult.nullExamples.length < 25) corpusResult.nullExamples.push(className);
            continue;
        }
        const candidate = parseCandidate(designSystem, className);
        if (!candidate) continue;

        /*
         * The variant prefix is stripped before lookup.
         *
         * `md:flex` and `flex` have the same reading — variants change the selector and the variant
         * bitmask, not the declarations — but the static table is keyed on the utility name, so
         * `md:flex` missed it and read as unpredictable. Forty-two of the corpus's 1,233 measurable
         * classes are variant-prefixed, and every one of them was a lookup failure rather than a
         * model failure.
         */
        const utilityName = className.slice(className.lastIndexOf(':') + 1);
        const predicted = candidate.kind === 'arbitrary'
            ? arbitraryPropertyReading(candidate)
            : candidate.kind === 'static'
              ? (staticReadings.get(utilityName) ?? null)
              : descriptors.has(candidate.root)
                ? predictReading(descriptors.get(candidate.root), utilityName)
                : undefined;

        corpusResult.measured++;
        if (sameReading(predicted, actual)) {
            corpusResult.agreed++;
            continue;
        }
        if (corpusResult.disagreed.length < 200) {
            corpusResult.disagreed.push({ className, root: candidate.root ?? null, predicted: readingKey(predicted ?? null), actual: readingKey(actual) });
        }
    }
}

// ---------------------------------------------------------------------------------------------
// Controls

/*
 * The planted-dirty control.
 *
 * One descriptor is corrupted on purpose and replayed. If the harness does not flag it, the harness
 * cannot flag anything, and the clean result above is worth nothing. The root chosen is whichever
 * one the registry exercises most, so the control runs against real traffic rather than a synthetic
 * class that might not be reached.
 */
const controlResult = { root: null, plantedProbes: 0, flagged: 0, proven: false };
{
    const classesByRoot = new Map();
    for (const className of registryClassNames) {
        const candidate = parseCandidate(designSystem, className);
        if (candidate?.kind !== 'functional' || !candidate.root) continue;
        if (!descriptors.has(candidate.root)) continue;
        if (readingOf(designSystem, className) === null) continue;
        const list = classesByRoot.get(candidate.root) ?? [];
        list.push(className);
        classesByRoot.set(candidate.root, list);
    }

    let bestRoot = null;
    for (const [root, list] of classesByRoot) {
        if (bestRoot === null || list.length > classesByRoot.get(bestRoot).length) bestRoot = root;
    }

    if (bestRoot !== null) {
        const original = descriptors.get(bestRoot);
        // Shift every order index by one and add a declaration. A wrong-but-plausible reading, not
        // an obviously broken one, because the check under test is exact structural comparison.
        const dirty = {
            root: original.root,
            typeList: original.typeList,
            readingByType: new Map(
                Array.from(original.readingByType, ([type, reading]) => [type, { order: reading.order.map((value) => value + 1), count: reading.count + 1 }]),
            ),
            fallbackReading: original.fallbackReading === null ? null : { order: original.fallbackReading.order.map((value) => value + 1), count: original.fallbackReading.count + 1 },
            readingByNamespace: new Map(
                Array.from(original.readingByNamespace, ([namespace, reading]) => [namespace, { order: reading.order.map((value) => value + 1), count: reading.count + 1 }]),
            ),
            readingByLiteral: new Map(
                Array.from(original.readingByLiteral, ([literal, reading]) => [literal, { order: reading.order.map((value) => value + 1), count: reading.count + 1 }]),
            ),
            emptyReadingWithModifier: original.emptyReadingWithModifier === null ? null : { order: original.emptyReadingWithModifier.order.map((value) => value + 1), count: original.emptyReadingWithModifier.count + 1 },
            emptyReadingWithThemedModifier: original.emptyReadingWithThemedModifier === null ? null : { order: original.emptyReadingWithThemedModifier.order.map((value) => value + 1), count: original.emptyReadingWithThemedModifier.count + 1 },
            readingByTypeWithModifier: new Map(
                Array.from(original.readingByTypeWithModifier, ([type, reading]) => [type, { order: reading.order.map((value) => value + 1), count: reading.count + 1 }]),
            ),
            fallbackReadingWithModifier: original.fallbackReadingWithModifier === null ? null : { order: original.fallbackReadingWithModifier.order.map((value) => value + 1), count: original.fallbackReadingWithModifier.count + 1 },
            readingByNamespaceWithModifier: new Map(
                Array.from(original.readingByNamespaceWithModifier, ([namespace, reading]) => [namespace, { order: reading.order.map((value) => value + 1), count: reading.count + 1 }]),
            ),
            readingByTypeWithThemedModifier: new Map(
                Array.from(original.readingByTypeWithThemedModifier, ([type, reading]) => [type, { order: reading.order.map((value) => value + 1), count: reading.count + 1 }]),
            ),
            fallbackReadingWithThemedModifier: original.fallbackReadingWithThemedModifier === null ? null : { order: original.fallbackReadingWithThemedModifier.order.map((value) => value + 1), count: original.fallbackReadingWithThemedModifier.count + 1 },
            readingByNamespaceWithThemedModifier: new Map(
                Array.from(original.readingByNamespaceWithThemedModifier, ([namespace, reading]) => [namespace, { order: reading.order.map((value) => value + 1), count: reading.count + 1 }]),
            ),
            emptyReading: original.emptyReading === null ? null : { order: original.emptyReading.order.map((value) => value + 1), count: original.emptyReading.count + 1 },
        };

        controlResult.root = bestRoot;
        for (const className of classesByRoot.get(bestRoot)) {
            const actual = readingOf(designSystem, className);
            if (actual === null) continue;
            controlResult.plantedProbes++;
            if (!sameReading(predictReading(dirty, className), actual)) controlResult.flagged++;
        }
        controlResult.proven = controlResult.plantedProbes > 0 && controlResult.flagged === controlResult.plantedProbes;
    }
}

/*
 * The end-to-end control, through the public API.
 *
 * Everything above reads `compileAstNodes`, which is internal and which upstream may rename.
 * `getClassOrder` is the stable public substitute: it answers "what is the relative position of
 * these classes" rather than "what did this class declare", so it is less diagnostic but it is not
 * exposed to the same rename. Sorting the registry by the predicted reading and by the engine's own
 * order, and comparing, checks the internal API against the public one.
 */
const publicApiResult = { ranked: 0, comparedPairs: 0, mismatched: 0, sample: [] };
{
    /*
     * A sample of the registry, strided so it spans the whole class list rather than its alphabetical
     * head. `getClassOrder` returns positions within the queried set, so the whole sample is queried
     * in one call and the positions read as a total order over it.
     */
    const stride = Math.max(1, Math.floor(registryClassNames.length / 3000));
    const sample = [];
    for (let index = 0; index < registryClassNames.length; index += stride) sample.push(registryClassNames[index]);

    const engineOrder = new Map(
        designSystem
            .getClassOrder(sample)
            .filter(([, position]) => position !== null)
            .map(([className, position]) => [className, position]),
    );

    // Only classes the descriptor model can predict at all. A class it declines to predict is not
    // evidence about the public API either way.
    const predicted = new Map();
    for (const className of sample) {
        if (!engineOrder.has(className)) continue;
        const candidate = parseCandidate(designSystem, className);
        if (!candidate) continue;
        const reading = candidate.kind === 'static'
            ? (staticReadings.get(className) ?? null)
            : (descriptors.has(candidate.root) ? predictReading(descriptors.get(candidate.root), className) : undefined);
        if (reading === null || reading === undefined) continue;
        predicted.set(className, reading);
    }

    const ranked = Array.from(predicted.keys());
    publicApiResult.ranked = ranked.length;

    /*
     * Every pair, not adjacent pairs.
     *
     * A first version walked adjacent entries of the registry list, which pairs classes that happen
     * to be alphabetical neighbours and says nothing about the sort. The order under test is a total
     * order, so the check is every pair of it: does the predicted reading order the pair the way the
     * engine does. Pairs whose predicted readings are identical are skipped, since those fall to the
     * variant and alphabetical tiebreaks this control does not model.
     */
    for (let leftIndex = 0; leftIndex < ranked.length; leftIndex++) {
        for (let rightIndex = leftIndex + 1; rightIndex < ranked.length; rightIndex++) {
            const left = ranked[leftIndex];
            const right = ranked[rightIndex];
            const leftReading = predicted.get(left);
            const rightReading = predicted.get(right);
            const comparison = compareReadings(leftReading, rightReading);
            if (comparison === 0) continue;
            publicApiResult.comparedPairs++;
            const engineSaysLeftFirst = engineOrder.get(left) < engineOrder.get(right);
            if (engineSaysLeftFirst !== comparison < 0) {
                publicApiResult.mismatched++;
                if (publicApiResult.sample.length < 20) {
                    publicApiResult.sample.push({
                        left,
                        right,
                        leftReading: readingKey(leftReading),
                        rightReading: readingKey(rightReading),
                        // `getClassOrder` returns BigInt positions, which JSON cannot serialize.
                        enginePositions: [String(engineOrder.get(left)), String(engineOrder.get(right))],
                    });
                }
            }
        }
    }
}

/*
 * Tailwind's own comparator, from `compile.ts`: first differing property index, then more
 * declarations first. Variants and the final alphabetical tiebreak are omitted because the control
 * that uses it skips pairs whose readings are identical, which is exactly the case those tiebreaks
 * decide.
 */
function compareReadings(left, right) {
    /*
     * An exhausted order sorts LAST, not first.
     *
     * A class whose order is empty declares nothing `PropertyOrder` ranks — `zoom-in-40` sets only
     * `--enter-scale` — and the engine puts those after every ranked utility. Returning -1 for the
     * shorter list instead put them first, and this control reported 365,174 mismatches out of
     * 4.8 million pairs while the model itself was fine. The bug was in the check, not the thing
     * being checked, which is exactly the failure mode a control is supposed to be immune to.
     */
    const length = Math.max(left.order.length, right.order.length);
    for (let index = 0; index < length; index++) {
        const leftValue = left.order[index];
        const rightValue = right.order[index];
        if (leftValue === rightValue) continue;
        if (leftValue === undefined) return 1;
        if (rightValue === undefined) return -1;
        return leftValue - rightValue;
    }
    return right.count - left.count;
}

// ---------------------------------------------------------------------------------------------
// Report

const readingCountHistogram = new Map();
for (const readings of sweepResult.readingsPerRoot.values()) {
    readingCountHistogram.set(readings.size, (readingCountHistogram.get(readings.size) ?? 0) + 1);
}

const disagreeingRoots = new Map();
for (const entry of [...registryResult.disagreed, ...sweepResult.disagreed]) {
    const list = disagreeingRoots.get(entry.root) ?? [];
    list.push(entry);
    disagreeingRoots.set(entry.root, list);
}

/*
 * What this design system contributes over a bare Tailwind install.
 *
 * Reported next to the volumes rather than instead of them, so a reader sees "37,641 registry, of
 * which 35 utility roots and 325 theme keys are this repository's" instead of a number that cannot
 * distinguish a repository from the framework it imports.
 */
const ownContributions = await surveyOwnContributions(designSystem, resolveRoot);

const report = {
    tailwindVersion,
    entryPoint: recordedEntryPoint,
    groundTruth: 'compileAstNodes (internal API; getClassOrder used as the public end-to-end control)',
    population: {
        registryClasses: registryClassNames.length,
        functionalRoots: roots.length,
        staticUtilities: staticNames.size,
        shapes: Shapes.length,
        ownUtilityRoots: ownContributions.ownUtilityRootCount,
        ownThemeKeys: ownContributions.ownThemeKeyCount,
        baselineRegistryClasses: ownContributions.baselineRegistryClasses,
    },
    ownContributions,
    registry: {
        measured: registryResult.measured,
        agreed: registryResult.agreed,
        disagreed: registryResult.disagreed.length,
        nullReading: registryResult.nullReading,
        unparsed: registryResult.unparsed,
    },
    corpus: {
        note: 'The corpus, not the registry. This is the population the documented null figure came from; a class with no reading is counted here and never scored as agreement.',
        classNameOccurrences: corpusResult.scanned,
        distinctClasses: corpusResult.distinct,
        measured: corpusResult.measured,
        agreed: corpusResult.agreed,
        disagreed: corpusResult.disagreed.length,
        nullReading: corpusResult.nullReading,
        nullExamples: corpusResult.nullExamples,
        examples: corpusResult.disagreed.slice(0, 12),
    },
    modifiers: {
        population: modifierClassNames.length,
        measured: modifierResult.measured,
        agreed: modifierResult.agreed,
        disagreed: modifierResult.disagreed.length,
        nullReading: modifierResult.nullReading,
        examples: modifierResult.disagreed.slice(0, 8),
    },
    sweep: {
        probed: sweepResult.probed,
        measured: sweepResult.measured,
        agreed: sweepResult.agreed,
        disagreed: sweepResult.disagreed.length,
        nullReading: sweepResult.nullReading,
        skippedUnregistered: sweepResult.skippedUnregistered,
    },
    readingsPerRootHistogram: Object.fromEntries(Array.from(readingCountHistogram).sort((left, right) => left[0] - right[0])),
    controls: {
        plantedDirty: controlResult,
        publicApi: publicApiResult,
    },
    disagreeingRoots: Array.from(disagreeingRoots, ([root, entries]) => ({
        root,
        failures: entries.length,
        examples: entries.slice(0, 8),
    })).sort((left, right) => right.failures - left.failures),
};

/*
 * Volume assertions.
 *
 * A run that measured almost nothing and agreed with all of it is indistinguishable, in the summary
 * numbers, from a run that measured everything. So the magnitudes are asserted rather than merely
 * reported: the registry is ~37,600 classes on both repos probed, the sweep is six figures, and the
 * controls have to have actually run. A run below these floors exits non-zero and its clean
 * agreement means nothing.
 *
 * The floors are deliberately loose. They are here to catch a loader that silently returned an empty
 * design system, not to pin exact counts that a Tailwind upgrade would legitimately move.
 */
const assertionFailures = [];
if (report.population.registryClasses < 20000) {
    assertionFailures.push('registry holds ' + report.population.registryClasses + ' classes; expected tens of thousands. The design system probably failed to load its @import graph.');
}
if (report.population.functionalRoots < 200) {
    assertionFailures.push('only ' + report.population.functionalRoots + ' functional roots; expected 300 or so.');
}
if (report.sweep.probed < 100000) {
    assertionFailures.push('sweep probed ' + report.sweep.probed + ' classes; expected six figures.');
}
if (report.sweep.measured < 50000) {
    assertionFailures.push('sweep measured only ' + report.sweep.measured + ' readings out of ' + report.sweep.probed + ' probes.');
}
if (propertyOrder.length < 300) {
    assertionFailures.push('the property order has ' + propertyOrder.length + ' entries; expected around 359. Arbitrary properties cannot be predicted without it.');
}
if (report.corpus.distinctClasses < 100) {
    assertionFailures.push('the corpus scan found only ' + report.corpus.distinctClasses + ' distinct classes; it probably scanned the wrong directory, so its null count means nothing.');
}
if (!report.controls.plantedDirty.proven) {
    assertionFailures.push('the planted-dirty control did not flag every probe, so this harness has not been shown to be able to fail.');
}

/*
 * Every functional root the registry knows has a descriptor.
 *
 * A volume floor cannot catch what this catches, and that is the whole reason it is written as a set
 * difference rather than as a count. 327 roots and 341 roots both clear "expected 300 or so"
 * comfortably, so the fourteen missing ones sat under a passing assertion for the life of this
 * table. The failure they produce downstream is a decline, which is indistinguishable in any summary
 * from the model's own declared boundary — the model declines 407 classes on purpose here — so the
 * gap is invisible on both sides at once unless the two populations are compared directly.
 *
 * Named per root rather than counted, because the point is which one is missing.
 */
const registryFunctionalRoots = Array.from(designSystem.utilities?.keys?.('functional') ?? [])
    .filter((root) => typeof root === 'string');
const rootsWithoutDescriptor = registryFunctionalRoots.filter((root) => !descriptors.has(root)).sort();
if (registryFunctionalRoots.length === 0) {
    assertionFailures.push('the registry reported no functional roots at all, so the coverage check below compared nothing and cannot have failed.');
} else if (rootsWithoutDescriptor.length > 0) {
    assertionFailures.push(
        rootsWithoutDescriptor.length +
            ' of ' +
            registryFunctionalRoots.length +
            ' registered functional roots have no descriptor, so a class naming one declines and reads as the model\'s boundary rather than as this gap: ' +
            rootsWithoutDescriptor.join(', '),
    );
}

/*
 * The design system is this repository's, and not bare Tailwind wearing its name.
 *
 * Every floor above is cleared by a stylesheet containing nothing but `@import "tailwindcss"`,
 * because Tailwind registers its built-in roots from JavaScript rather than from CSS: 23,286
 * registry classes, 301 functional roots, six figures of sweep probes, a full property order and a
 * proven planted control. That run was measured, and it passed. What it lacks is the `@utility` and
 * `@theme` content that makes a design system per-repository at all, which is precisely what the
 * table generated from it would claim to describe.
 *
 * This is the same defect a219c8f found in the candidate parser's own volume assertion, and it is
 * the bug this whole port exists to fix pointing the other way: a table describing one repository
 * while claiming to describe Tailwind, and a generator describing bare Tailwind while claiming to
 * describe a repository, are one error with two signs. Both look completely clean in the summary.
 *
 * Counted, not named. The floors are 1 because the honest claim is "this system contributes
 * something of its own", and any number above that is a fact about one repository rather than about
 * repositories: ahra contributes 35 roots and 325 theme keys, the independent fixture contributes 2
 * and 4, and both are legitimate design systems. A floor tuned to ahra would refuse the fixture.
 */
if (report.ownContributions.ownUtilityRootCount < 1 && report.ownContributions.ownThemeKeyCount < 1) {
    assertionFailures.push(
        'this design system registers nothing a bare Tailwind install does not: ' +
            report.population.registryClasses + ' registry classes against a baseline of ' +
            report.ownContributions.baselineRegistryClasses + ', 0 own @utility roots and 0 own @theme keys. ' +
            'The entry point loaded, and its @import graph reached no repository stylesheet, so this table ' +
            'describes the framework while claiming to describe a repository.',
    );
}
report.assertions = { passed: assertionFailures.length === 0, failures: assertionFailures };

process.stdout.write(JSON.stringify(report, null, 2) + '\n');

if (assertionFailures.length > 0) {
    process.stderr.write('extract.mjs: volume assertions failed, so the agreement numbers above mean nothing:\n');
    for (const failure of assertionFailures) process.stderr.write('  - ' + failure + '\n');
    process.exitCode = 4;
}

if (jsonOutputPath) {
    const table = {
        tailwindVersion,
        entryPoint: recordedEntryPoint,
        roots: Array.from(descriptors.values(), (descriptor) => ({
            root: descriptor.root,
            typeList: descriptor.typeList,
            readingByType: Object.fromEntries(Array.from(descriptor.readingByType, ([type, reading]) => [type, reading])),
            fallback: descriptor.fallbackReading,
            readingByNamespace: Object.fromEntries(Array.from(descriptor.readingByNamespace, ([namespace, reading]) => [namespace, reading])),
            readingByLiteral: Object.fromEntries(Array.from(descriptor.readingByLiteral, ([literal, reading]) => [literal, reading])),
            empty: descriptor.emptyReading,
            emptyWithModifier: descriptor.emptyReadingWithModifier,
            emptyWithThemedModifier: descriptor.emptyReadingWithThemedModifier,
            readingByTypeWithModifier: Object.fromEntries(Array.from(descriptor.readingByTypeWithModifier, ([type, reading]) => [type, reading])),
            fallbackWithModifier: descriptor.fallbackReadingWithModifier,
            readingByNamespaceWithModifier: Object.fromEntries(Array.from(descriptor.readingByNamespaceWithModifier, ([namespace, reading]) => [namespace, reading])),
            readingByTypeWithThemedModifier: Object.fromEntries(Array.from(descriptor.readingByTypeWithThemedModifier, ([type, reading]) => [type, reading])),
            fallbackWithThemedModifier: descriptor.fallbackReadingWithThemedModifier,
            readingByNamespaceWithThemedModifier: Object.fromEntries(Array.from(descriptor.readingByNamespaceWithThemedModifier, ([namespace, reading]) => [namespace, reading])),
        })).sort((left, right) => (left.root < right.root ? -1 : 1)),
        statics: Object.fromEntries(Array.from(staticReadings).sort((left, right) => (left[0] < right[0] ? -1 : 1))),
    };
    NodeFileSystem.writeFileSync(jsonOutputPath, JSON.stringify(table, null, 2) + '\n');
}
