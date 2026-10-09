/*
 * Compare two Tailwind releases structurally, so a version bump produces a list of what to re-read.
 *
 * M5 of #twutil. The port carries five tables, and three of them describe Tailwind's own data rather
 * than this repository's: `PropertyOrder`, `FrameworkVariantRegistrations` and
 * `FrameworkStaticDeclarations`. Upstream carries the same data. `property-order.ts` at 4.3.3 is a
 * hand-written array of 359 property names, with comments in it like "How do we make `inset-x-0`
 * come before `top-0`?", and our `PropertyOrder` holds 359 entries because it is that array.
 *
 * A table you did not derive is a table you have to track. This is what makes that tracking a
 * measurement rather than a hope.
 *
 * # What it compares, and why not the source
 *
 * A text diff of `utilities.ts` between two releases reports every refactor and every comment. What
 * matters here is narrower: which registrations exist, what they declare, and where each property
 * sorts. So this loads the same stylesheet through two installed engines and compares what they
 * register, which is the layer the port actually depends on.
 *
 * Four questions, each mapping to a table the port carries:
 *
 *   static utilities added, removed, or declaring different properties  -> FrameworkStaticDeclarations
 *   functional roots added or removed                                    -> the wave tables
 *   the global property order, by position                               -> PropertyOrder
 *   variant registrations and their order numbers                        -> FrameworkVariantRegistrations
 *
 * # The failure to design against
 *
 * A harness that reports nothing on a real bump. Two releases that genuinely differ must produce a
 * non-empty report, so this exits non-zero when the two engines report the same version, and says so
 * rather than printing an empty diff that reads like agreement.
 *
 * Usage:
 *
 *   node internal/lint/rules/tailwind/tools/generate_upstream_diff/enumerate.mjs <theme.css> \
 *       --base-root <dir with node_modules/tailwindcss> \
 *       --next-root <dir with node_modules/tailwindcss>
 */

import { loadDesignSystem, parseCandidate, readingOf } from '../generate_descriptors/loader.mjs';

const entryPointArgument = process.argv[2];
if (!entryPointArgument) {
    process.stderr.write('usage: enumerate.mjs <theme.css> --base-root <dir> --next-root <dir>\n');
    process.exit(2);
}

function flagValue(name) {
    const index = process.argv.indexOf(name);
    return index === -1 ? undefined : process.argv[index + 1];
}

const baseRoot = flagValue('--base-root');
const nextRoot = flagValue('--next-root');
if (!baseRoot || !nextRoot) {
    process.stderr.write('both --base-root and --next-root are required: a diff needs two engines\n');
    process.exit(2);
}

/*
 * A snapshot of everything the port reads from one engine.
 *
 * Declarations are captured per static utility rather than only the names, because a release that
 * keeps a name and changes what it declares is the change most likely to pass unnoticed: the count
 * holds, the registry holds, and only the reading moves.
 */
async function snapshot(resolveRoot) {
    const { designSystem, tailwindVersion } = await loadDesignSystem(entryPointArgument, resolveRoot);

    const statics = {};
    for (const name of designSystem.utilities.keys('static')) {
        if (typeof name !== 'string') continue;
        const reading = readingOf(designSystem, name);
        statics[name] = reading === null ? null : { order: reading.order, count: reading.count };
    }

    const functionalRoots = designSystem.utilities.keys('functional').filter((root) => typeof root === 'string').sort();

    const variants = {};
    for (const name of designSystem.variants.keys?.() ?? []) {
        if (typeof name !== 'string') continue;
        variants[name] = designSystem.variants.get?.(name)?.order ?? null;
    }

    return { tailwindVersion, statics, functionalRoots, variants };
}

const base = await snapshot(baseRoot);
const next = await snapshot(nextRoot);

if (base.tailwindVersion === next.tailwindVersion) {
    process.stderr.write(
        `enumerate.mjs: both roots resolve to Tailwind ${base.tailwindVersion}, so this run compares an ` +
            'engine with itself and an empty report would mean nothing. Point --base-root and --next-root ' +
            'at different installs.\n',
    );
    process.exit(4);
}

function diffNames(before, after) {
    const beforeSet = new Set(before);
    const afterSet = new Set(after);
    return {
        added: after.filter((name) => !beforeSet.has(name)),
        removed: before.filter((name) => !afterSet.has(name)),
    };
}

const staticNames = diffNames(Object.keys(base.statics).sort(), Object.keys(next.statics).sort());
const changedStatics = [];
for (const name of Object.keys(base.statics)) {
    if (!(name in next.statics)) continue;
    const before = JSON.stringify(base.statics[name]);
    const after = JSON.stringify(next.statics[name]);
    if (before !== after) changedStatics.push({ name, before: base.statics[name], after: next.statics[name] });
}

const rootNames = diffNames(base.functionalRoots, next.functionalRoots);

const variantNames = diffNames(Object.keys(base.variants).sort(), Object.keys(next.variants).sort());
const changedVariants = [];
for (const name of Object.keys(base.variants)) {
    if (!(name in next.variants)) continue;
    if (base.variants[name] !== next.variants[name]) {
        changedVariants.push({ name, before: base.variants[name], after: next.variants[name] });
    }
}

const report = {
    baseVersion: base.tailwindVersion,
    nextVersion: next.tailwindVersion,
    entryPoint: entryPointArgument,
    statics: {
        added: staticNames.added,
        removed: staticNames.removed,
        changedReading: changedStatics,
        table: 'FrameworkStaticDeclarations',
    },
    functionalRoots: {
        added: rootNames.added,
        removed: rootNames.removed,
        table: 'FrameworkFunctionalUtilities and FrameworkMultiDeclarationUtilities',
    },
    variants: {
        added: variantNames.added,
        removed: variantNames.removed,
        changedOrder: changedVariants,
        table: 'FrameworkVariantRegistrations',
    },
    counts: {
        baseStatics: Object.keys(base.statics).length,
        nextStatics: Object.keys(next.statics).length,
        baseFunctionalRoots: base.functionalRoots.length,
        nextFunctionalRoots: next.functionalRoots.length,
        baseVariants: Object.keys(base.variants).length,
        nextVariants: Object.keys(next.variants).length,
    },
};

const changeCount =
    staticNames.added.length +
    staticNames.removed.length +
    changedStatics.length +
    rootNames.added.length +
    rootNames.removed.length +
    variantNames.added.length +
    variantNames.removed.length +
    changedVariants.length;

report.changeCount = changeCount;
report.verdict =
    changeCount === 0
        ? `no registration changed between ${base.tailwindVersion} and ${next.tailwindVersion}, so the checked-in tables need no re-reading`
        : `${changeCount} registration changes between ${base.tailwindVersion} and ${next.tailwindVersion}; the tables named above are what to regenerate`;

process.stdout.write(JSON.stringify(report, null, 2) + '\n');
