/*
 * The engine side of the Tier 2 differential: `{order, count}` for every class, on every design
 * system, as the shipped Tailwind computes it.
 *
 * This is not another component fixture. Six components already carry their own engine-verified
 * corpora and each of those hands Go a value it has already decoded: `descriptor_fixtures.json`
 * supplies a parsed `{kind, root, value, modifier}` and asks only whether `Table.Lookup` maps it to
 * the right reading. That is the right shape for testing a lookup and it cannot see the seam. A
 * class is a *string*, and turning the string into the candidate is `ParseCandidate`'s job, so a
 * suite that starts from a candidate has tested both halves and never the join between them.
 *
 * `segment` is the recorded precedent: it was verified exclusively through `InferDataType`, which
 * launders its errors, and mutations to it changed nothing observable. A component reachable only
 * through its caller is not tested, and neither is a *seam* reachable only through a fixture that
 * has already crossed it.
 *
 * So this file captures the one thing no component fixture can: the class name, and the answer.
 * What happens in between is entirely the Go side's problem, which is the point.
 *
 * Usage:
 *
 *   node internal/lint/rules/tailwind/tools/generate_class_order/enumerate.mjs <systems.json>
 *
 * Writes the fixture to stdout and its progress to stderr.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodePath from 'node:path';
import { recordBeside } from '../corpus.mjs';
import { loadDesignSystem, parseCandidate, readingOf } from '../generate_descriptors/loader.mjs';

const systemsArgument = process.argv[2];
if (!systemsArgument) {
    process.stderr.write('usage: enumerate.mjs <systems.json>\n');
    process.exit(2);
}

const systemDefinitions = JSON.parse(NodeFileSystem.readFileSync(NodePath.resolve(systemsArgument), 'utf8'));
if (!Array.isArray(systemDefinitions) || systemDefinitions.length < 2) {
    // Two is the floor rather than a preference. One design system is not a population: the entire
    // argument for this port is that two repositories on the same Tailwind resolve differently, and
    // a harness that can only see one of them cannot observe the defect it exists to catch.
    process.stderr.write('enumerate.mjs: at least two design systems are required; one is not a population\n');
    process.exit(2);
}

/*
 * The registry, expanded.
 *
 * `getClassList()` returns entries that are either a bare class name or a `[name, meta]` pair whose
 * meta carries the modifiers the engine will accept. The modifiers matter and are not decoration:
 * the descriptor model's third axis is the modifier, with three states rather than two, and a
 * registry enumerated without them exercises exactly one of the three. So a class advertising
 * modifiers contributes the bare class plus one sample per axis.
 *
 * Sampled rather than exhausted on purpose. A colour utility advertises every theme key as a
 * modifier and the cross product runs to millions, none of which say anything the three axes do not:
 * `/25`, `/[0.5]` and `/[var(--a)]` are one bucket, measured in Phase 0. The three probes chosen
 * here are one per bucket, and the bucket boundary is the thing under test.
 *
 * Expanding the modifiers is also what makes the registry a population with nulls in it. The bare
 * registry has none by construction, every entry being a class the engine advertised, and the
 * documented null figure therefore had to come from the corpus. Advertised modifiers are not the
 * same guarantee: `border` advertises modifiers and `border/50` compiles to nothing, 30 classes per
 * system doing this. The engine advertises a modifier it then refuses, so the expanded registry
 * carries the case a differential must not score as agreement — in the population where nobody
 * expected to find it.
 */
function expandRegistry(designSystem) {
    const classNames = [];
    const seen = new Set();

    function add(className) {
        if (seen.has(className)) return;
        seen.add(className);
        classNames.push(className);
    }

    for (const entry of designSystem.getClassList()) {
        const name = Array.isArray(entry) ? entry[0] : entry;
        const meta = Array.isArray(entry) ? entry[1] : null;
        add(name);

        const declaredModifiers = meta?.modifiers;
        if (!Array.isArray(declaredModifiers) || declaredModifiers.length === 0) continue;

        // One probe per modifier axis. `50` is the alpha axis; a declared theme-key modifier is the
        // themed axis; the bare class above is the absent axis. `[var(--a)]` is added because an
        // arbitrary modifier reaches the alpha axis by a different route through the parser, and
        // the seam under test is the parser.
        add(name + '/50');
        add(name + '/[var(--a)]');
        const themedModifier = declaredModifiers.find((modifier) => typeof modifier === 'string' && !/^\d+$/.test(modifier));
        if (themedModifier !== undefined) add(name + '/' + themedModifier);
    }

    return classNames;
}

/*
 * The corpus: the classes the repository actually writes.
 *
 * Carried alongside the registry rather than instead of it, because they are different populations
 * and only one of them contains the case this harness was built to refuse to score. The registry has
 * no null readings by construction, every entry in it being a class the engine advertised. The
 * corpus does: 10 of 1,208 distinct classes on ahra return null, and those ten are precisely where a
 * differential that reads silence as agreement would be blind.
 *
 * Scanned the same way `generate_descriptors/extract.mjs` scans, so the two populations are
 * comparable and the null figure this harness asserts against is the same figure that one measured.
 */
function scanCorpus(entryPoint) {
    // `<repo>/app/_theme/styles/theme.css` -> `<repo>`.
    const corpusRoot = NodePath.dirname(NodePath.dirname(NodePath.dirname(NodePath.dirname(entryPoint))));
    const distinctClasses = new Set();
    let occurrences = 0;

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
                    occurrences++;
                    distinctClasses.add(className);
                }
            }
        }
    }

    scanDirectory(corpusRoot, 0);
    return { corpusRoot, occurrences, classNames: Array.from(distinctClasses).sort() };
}

/*
 * The registration tables the Go parser needs.
 *
 * `ParseCandidate` takes a `DesignSystem` and asks it four questions, and the answers are per
 * repository. They are captured here, in this fixture, rather than borrowed from
 * `candidate_fixtures.json`: that file belongs to the candidate parser's own suite and its shape is
 * that author's to change. A harness whose population silently depends on another component's
 * fixture breaks when that component is refactored, and the break looks like a differential finding.
 */
function registrationTables(designSystem) {
    /*
     * Utility roots and their registration kinds, read off the engine's own registry.
     *
     * Reached through `utilities.utilities`, the backing Map from root to its declarations, rather
     * than reconstructed from `getClassList()`. The registry advertises *names* and the parser asks
     * about *roots*: `bg-red-500` is the entry and `bg` is the root, and recovering one from the
     * other means re-implementing the split the parser is under test for.
     *
     * A root can hold both kinds. `flex` is a static utility and also a functional one, so the value
     * is a list rather than a single kind, and collapsing it would make `HasUtility(root, kind)`
     * answer for the wrong registration on every root that carries two.
     */
    const utilityRoots = {};
    for (const [root, declarations] of designSystem.utilities.utilities) {
        const kinds = new Set();
        for (const declaration of Array.isArray(declarations) ? declarations : [declarations]) {
            if (declaration?.kind === 'static' || declaration?.kind === 'functional') kinds.add(declaration.kind);
        }
        if (kinds.size > 0) utilityRoots[root] = Array.from(kinds).sort();
    }

    if (Object.keys(utilityRoots).length === 0) {
        // Loud rather than empty. An empty utility table makes every class parse as nothing, the Go
        // side declines everything, and the run reports total silence — which this harness fails on
        // by design, but it would fail naming the wrong cause.
        process.stderr.write('enumerate.mjs: the utility registry could not be read; the parser tables would be empty\n');
        process.exit(3);
    }

    /*
     * The variant table, including both compound bitmasks.
     *
     * `compounds` is what a variant produces and `compoundsWith` is what it accepts, and they are
     * different numbers on the same variant: `not` is `2` and `3`. Capturing one and inferring the
     * other would make every compound-variant parse agree for the wrong reason.
     */
    const variantRoots = {};
    for (const [root, variant] of designSystem.variants.entries()) {
        variantRoots[root] = {
            kind: variant?.kind ?? 'static',
            compounds: variant?.compounds ?? 0,
            compoundsWith: variant?.compoundsWith ?? 0,
        };
    }

    // `theme.prefix` is `null` when unset and the parser's contract is an empty string, so the
    // coercion happens here rather than being left for each consumer to get right separately.
    return { prefix: designSystem.theme?.prefix ?? '', utilityRoots, variantRoots };
}

const systems = [];
for (const definition of systemDefinitions) {
    process.stderr.write(`enumerate.mjs: loading ${definition.name}\n`);
    const { designSystem, tailwindVersion, entryPoint, recordedEntryPoint } = await loadDesignSystem(definition.path);

    const registryClassNames = expandRegistry(designSystem);
    const corpus = scanCorpus(entryPoint);

    /*
     * The cases, with their readings interned.
     *
     * 95,136 registry classes carry 379 distinct readings between them, so the readings become a
     * deduplicated array and each case holds an index into it. Written out flat the fixture is 20 MB
     * and mostly repetitions of `{"order":[199,200],"count":2}`; interned it is a twentieth of that
     * and holds exactly the same answers.
     *
     * A null reading gets an index too, rather than a missing field. `readingIndex: -1` is the
     * engine having no answer, and it is a value in the same column as every other answer instead of
     * an absence a reader has to notice. This harness exists to refuse to score silence as
     * agreement, and it would be a poor start to spell silence as a field that is not there.
     */
    const readingKeys = new Map();
    const readings = [];
    function internReading(reading) {
        if (reading === null) return -1;
        const key = `[${reading.order.join(',')}]#${reading.count}`;
        const existing = readingKeys.get(key);
        if (existing !== undefined) return existing;
        const index = readings.length;
        readings.push({ order: reading.order, count: reading.count });
        readingKeys.set(key, index);
        return index;
    }

    const registryCases = [];
    const corpusCases = [];
    let registryNullCount = 0;
    let corpusNullCount = 0;
    const corpusNullExamples = [];

    for (const className of registryClassNames) {
        const reading = readingOf(designSystem, className);
        if (reading === null) registryNullCount++;
        registryCases.push({ className, readingIndex: internReading(reading) });
    }

    for (const className of corpus.classNames) {
        const reading = readingOf(designSystem, className);
        if (reading === null) {
            corpusNullCount++;
            if (corpusNullExamples.length < 25) corpusNullExamples.push(className);
        }
        corpusCases.push({ className, readingIndex: internReading(reading) });
    }

    /*
     * The public-API cross-check, carried as its own population.
     *
     * `readingOf` uses `compileAstNodes`, which is internal, and it is the right ground truth
     * because it answers "what did this class declare" rather than "where did it sort". The public
     * `getClassOrder` answers a weaker question and is captured anyway, over a sample, so that a
     * rename or a semantic change upstream shows up as a failing gate instead of silently changing
     * every number in this fixture at once.
     */
    const crossCheckSample = registryClassNames.filter((_, index) => index % 37 === 0).slice(0, 2000);
    const classOrder = designSystem.getClassOrder(crossCheckSample);
    const publicOrder = [];
    for (const [className, order] of classOrder) {
        publicOrder.push({ className, order: order === null ? null : String(order) });
    }

    process.stderr.write(
        `  ${definition.name}: ${registryClassNames.length} registry, ${corpus.classNames.length} corpus, `
        + `${registryNullCount} registry nulls, ${corpusNullCount} corpus nulls\n`,
    );

    systems.push({
        name: definition.name,
        entryPoint: recordedEntryPoint,
        tailwindVersion,
        corpusRoot: recordBeside(definition.path, corpus.corpusRoot),
        counts: {
            registryClasses: registryClassNames.length,
            corpusOccurrences: corpus.occurrences,
            corpusClasses: corpus.classNames.length,
            registryNullReadings: registryNullCount,
            corpusNullReadings: corpusNullCount,
            distinctReadings: readings.length,
        },
        corpusNullExamples,
        registration: registrationTables(designSystem),
        publicOrder,
        readings,
        registryCases,
        corpusCases,
    });
}

/*
 * The divergence assertion, and the axis it turned out to live on.
 *
 * `f7d1d8d` is the commit to copy: it took the claim the whole port rests on and made it an
 * assertion that fails if it ever becomes trivially true. The claim is that a table generated from
 * one repository is wrong for another, and this measures where that is true.
 *
 * It is not where reading the theme result suggests. Theme resolution diverges on *values*: 744
 * entries against 748, with 22 shared keys carrying different values. A reading is `{order, count}`,
 * and order is the index of the declared *property*, so a shared class declaring the same property
 * with a different value reads identically in both. Measured: 95,134 classes are in both registries
 * and every one of them reads the same.
 *
 * The divergence is in *membership*. 462 registry classes exist in exactly one of the two systems,
 * because a `--color-brand-*` token in one repository's `@theme` is what makes `accent-brand` a
 * class at all. A table generated against ahra has no row that answers `accent-brand`, so it
 * declines a class the engine compiles, which is a silent-side defect rather than a wrong reading.
 *
 * That is the sharper claim and it is the one asserted: both axes are recorded, so a future Tailwind
 * that starts diverging readings shows up as a number moving off zero rather than as a fixture that
 * quietly still passes.
 */
const divergence = { sharedClasses: 0, divergentReadings: 0, uniqueToOneSystem: 0, readingExamples: [], membershipExamples: [] };
if (systems.length >= 2) {
    const [left, right] = systems;
    const registryOf = (system) => {
        const byName = new Map();
        for (const aCase of system.registryCases) {
            const reading = aCase.readingIndex < 0 ? null : system.readings[aCase.readingIndex];
            byName.set(aCase.className, reading === null ? 'null' : `[${reading.order.join(',')}]#${reading.count}`);
        }
        return byName;
    };

    const leftRegistry = registryOf(left);
    const rightRegistry = registryOf(right);

    for (const [className, leftKey] of leftRegistry) {
        if (!rightRegistry.has(className)) {
            divergence.uniqueToOneSystem++;
            if (divergence.membershipExamples.length < 25) {
                divergence.membershipExamples.push({ className, presentIn: left.name });
            }
            continue;
        }
        divergence.sharedClasses++;
        const rightKey = rightRegistry.get(className);
        if (leftKey === rightKey) continue;
        divergence.divergentReadings++;
        if (divergence.readingExamples.length < 25) {
            divergence.readingExamples.push({ className, left: leftKey, right: rightKey });
        }
    }

    for (const className of rightRegistry.keys()) {
        if (leftRegistry.has(className)) continue;
        divergence.uniqueToOneSystem++;
        if (divergence.membershipExamples.length < 25) {
            divergence.membershipExamples.push({ className, presentIn: right.name });
        }
    }
}

process.stderr.write(
    `enumerate.mjs: ${divergence.sharedClasses} classes in both registries, `
    + `${divergence.divergentReadings} read differently, `
    + `${divergence.uniqueToOneSystem} present in only one system\n`,
);

process.stdout.write(JSON.stringify({
    tailwindVersion: systems[0].tailwindVersion,
    groundTruth: 'compileAstNodes',
    divergence,
    systems,
}, null, 0) + '\n');
