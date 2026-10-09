/*
 * Ask the real Tailwind 4.3.3 engine what every static utility declares, and emit it as JSON.
 *
 * This is the first component of #twutil, the port that removes `descriptor_base_table.go`. A static
 * utility takes no value, so its reading is a constant, and the constant is a walk over the
 * declarations it compiles to. That makes it the cheapest half of `createUtilities` to port and the
 * one that proves the shape for the functional half.
 *
 * # Generated rather than transcribed, and the reason is not effort
 *
 * `utilities.ts` at the pinned tag holds 485 `staticUtility(...)` calls, 357 of them literal
 * `[[property, value]]` pairs. Copying them into Go by hand is 485 chances to transpose a property
 * name into something that still compiles and still looks right, and the failure would be a wrong
 * reading rather than an error. Every other component in this slice generates its fixtures against
 * the engine for exactly that reason: `generate_data_type`, `generate_css_parser` and
 * `generate_theme` all ask rather than transcribe.
 *
 * The source count is also not the registry count, which is the second reason. `staticUtility` is
 * one of several registration paths and the source-side grep undercounts: the registry reports 917
 * static utilities where the source shows 485 calls. The source says what upstream wrote; the
 * registry says what it registered, and the registry is what the linter has to agree with.
 *
 * # Ground truth
 *
 * `designSystem.compileAstNodes(candidate)` returns the compiled tree, whose node shape is the one
 * `internal/tailwind/ast.go` already models: rules holding declarations, with `property` and `value`
 * on each declaration. It is an internal API and upstream may rename it, which is why the emitted
 * file records the version it was generated against and why `readingOf` is carried alongside as an
 * independent check: if the tree and the reading ever disagree, that is a finding rather than noise.
 *
 * Usage:
 *
 *   node internal/lint/rules/tailwind/tools/generate_statics/enumerate.mjs <theme.css> [--resolve-root <dir>] > statics.json
 */

import { loadDesignSystem, parseCandidate, readingOf } from '../generate_descriptors/loader.mjs';

const entryPointArgument = process.argv[2];
if (!entryPointArgument) {
    process.stderr.write('usage: enumerate.mjs <theme.css> [--resolve-root <dir>]\n');
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

/*
 * The registry's own enumeration, not the class list.
 *
 * `getClassList()` enumerates classes and a utility that advertises none is invisible to it, which
 * is the defect commits 33e12df and 6347b07 fixed on the functional side. Reading statics from the
 * class list would reintroduce it here, one component later, in a file whose whole job is to be the
 * complete registration set.
 */
const staticNames = designSystem.utilities.keys('static').filter((name) => typeof name === 'string');

/*
 * Declarations are read off the compiled tree rather than off the source.
 *
 * A static utility can compile to a bare declaration list, to a rule holding one, or to several
 * rules: `sr-only` is nine declarations and `space-y-*` wraps its body in a `:where(...)` selector.
 * The walk below flattens to the sequence PropertySort would visit, which is breadth-first over the
 * container kinds it descends into, so the emitted `declarations` array is directly comparable to
 * what the Go side produces rather than requiring the consumer to re-derive an order.
 */
function flattenDeclarations(nodes) {
    const declarations = [];
    const queue = [...nodes];
    while (queue.length > 0) {
        const node = queue.shift();
        if (!node) continue;
        if (node.kind === 'declaration') {
            declarations.push({
                property: node.property,
                value: node.value ?? null,
                valuePresent: node.value !== undefined,
                important: node.important === true,
            });
            continue;
        }
        // Matching PropertySort: rules and at-rules are descended into, and nothing else is. The
        // `@property` bodies this deliberately skips are why `shadow-[--x]` prints 39 declarations
        // and reads `{order:[315,316], count:2}`.
        if (node.kind === 'rule' || node.kind === 'at-rule') {
            for (const child of node.nodes ?? []) queue.push(child);
        }
    }
    return declarations;
}

const entries = [];
const unresolved = [];

for (const name of staticNames) {
    const candidate = parseCandidate(designSystem, name);
    if (!candidate || candidate.kind !== 'static') {
        // Registered static but does not parse as one. Recorded rather than dropped: a name in the
        // registry that the parser reads differently is exactly the class of disagreement this port
        // exists to surface.
        unresolved.push({ name, reason: candidate ? 'parses as ' + candidate.kind : 'unparsed' });
        continue;
    }

    let nodes;
    try {
        nodes = designSystem.compileAstNodes(candidate);
    }
    catch (error) {
        unresolved.push({ name, reason: 'compileAstNodes threw: ' + String(error?.message ?? error).slice(0, 80) });
        continue;
    }
    if (!nodes || nodes.length === 0) {
        unresolved.push({ name, reason: 'compiled to nothing' });
        continue;
    }

    const declarations = flattenDeclarations([nodes[0].node]);
    const reading = readingOf(designSystem, name);

    entries.push({
        name,
        declarations,
        // Carried as an independent check rather than as the answer. The Go side computes its own
        // reading by walking the declarations; if the two disagree, the walk is wrong or the
        // flattening is, and the consumer can say which.
        reading: reading === null ? null : { order: reading.order, count: reading.count },
    });
}

entries.sort((left, right) => (left.name < right.name ? -1 : left.name > right.name ? 1 : 0));

process.stdout.write(JSON.stringify({
    tailwindVersion,
    entryPoint: recordedEntryPoint,
    groundTruth: 'compileAstNodes (internal API), with readingOf as an independent check',
    registeredStatics: staticNames.length,
    emitted: entries.length,
    unresolved,
    statics: entries,
}, null, 2) + '\n');
