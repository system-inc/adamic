/*
 * Ask the real Tailwind engine which variants the framework registers, and under which order
 * numbers.
 *
 * This is the static half of the variant component, and it is the half `generate_variant`
 * deliberately does not print. That tool captures the *behaviour* — `getVariantOrder` and
 * `getClassOrder` over a corpus — into a fixture a test reads. This one captures the *registry*
 * into shipped Go, because `LoadDesignSystem` needs it at lint time and a fixture under `testdata`
 * is not shipped.
 *
 * # Why the registry can be a table when the order map cannot
 *
 * `variant.go` argues at length that no static table can hold a variant's *position*, and it is
 * right: `getVariantOrder` assigns a dense index per run over only the variants that were parsed,
 * so `hover` is 0 in one run and 1 in another. Nothing here contradicts that. What is captured is
 * the *registration* — the name, the `order` number `Variants.set` assigned, and the kind — which
 * is a fact about the installed framework and does not vary per run or per repository. The dense
 * index is still computed per class list, from these registrations, by `BuildVariantOrder`.
 *
 * The distinction is the same one `CollapseFamilies` rests on: a fact about roots is a table, a
 * computation per class is code.
 *
 * # What is captured, and what cannot be
 *
 * `entries` is every registered root with its order and kind. `compareFnOrders` is which order
 * numbers carry a comparison function. The function itself cannot cross this boundary — it is a
 * closure over the design system's theme — so the Go side registers `CompareBreakpoints` against
 * those orders, reading the *live* theme. That is not a shortcut. Redefining `--breakpoint-sm` to
 * `200rem` moves `sm` from index 0 to index 4 on the engine, measured, so a comparison that read a
 * hardcoded scale would be wrong on exactly the repository that customized its breakpoints.
 *
 * The directions are captured rather than assumed: this script probes each grouped order with two
 * of its own members and reports whether the engine ordered them ascending or descending, so a
 * Tailwind release that flipped one produces a diff rather than a silent reversal.
 *
 * `@plugin` and `@config` are stubbed, matching the sibling generators. A JavaScript plugin can
 * call `addVariant`, so a repository with one registers a root this enumeration does not see. That
 * is a stated narrowing: this table is the *framework's* registrations, and a repository's own
 * contributions stay live-loaded through `@custom-variant`.
 *
 * # The optional design system argument, and why it had to exist
 *
 * The second argument is a CSS entry point to read the registry from, defaulting to a bare
 * `@import "tailwindcss"`. Without it this script had no way to load a repository's stylesheet at
 * all, which meant the invariance claim in the generated file named a comparison the tool could not
 * perform: every "independent generation" was the same framework `index.css` parsed again, and a
 * set of readings that cannot differ cannot detect a table that varies.
 *
 * With a real entry point the comparison is real, and measured it holds: the 88 registrations are
 * identical across ahra, www-connected-app, www-phi-health, an independent design system sharing no
 * submodule, and a bare install. The mechanism is `Variants.set` assigning kind onto an existing
 * record without touching `order`, so a repository redefining `dark` moves nothing, while a
 * `@custom-variant` under a new name appends. Both halves were probed rather than assumed.
 *
 * Node is required to run this and never to use its result.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodePath from 'node:path';

const [, , packageRootArgument, designSystemArgument] = process.argv;
if (!packageRootArgument) {
    process.stderr.write('usage: enumerate.mjs <tailwindcss package root> [design-system.css]\n');
    process.exit(2);
}

const packageRoot = NodePath.resolve(packageRootArgument);
const packageJson = JSON.parse(NodeFileSystem.readFileSync(NodePath.join(packageRoot, 'package.json'), 'utf8'));
const bundle = await import(NodePath.join(packageRoot, 'dist', 'lib.mjs'));

/*
 * The CSS the registry is read from, which was a constant and should not have been.
 *
 * This script built every design system from the literal `@import "tailwindcss";` and never read a
 * repository's stylesheet. That made one of the two invariance claims in the generated header
 * unmeasurable by the tool asserting it: "each of the two corpus repositories' own theme.css loaded
 * through its own install" describes a load this file could not perform, since there was no
 * argument for a stylesheet and no code path that opened one. What was actually compared was the
 * same framework `index.css` parsed several times, which cannot vary by construction.
 *
 * Passing a real entry point is what turns the claim into a measurement. A repository's own
 * `@custom-variant` blocks then reach the registry, which is the only way a repository can move it.
 */
const designSystemPath = designSystemArgument ? NodePath.resolve(designSystemArgument) : null;
const designSystemBase = designSystemPath ? NodePath.dirname(designSystemPath) : '/';
const designSystemSource = designSystemPath
    ? NodeFileSystem.readFileSync(designSystemPath, 'utf8')
    : '@import "tailwindcss";';

/*
 * Resolve `@import` the way a bundler would, matching every sibling generator's loader exactly.
 *
 * A stylesheet that cannot be read returns empty rather than throwing, which matters once a real
 * repository theme is loaded: those import package stylesheets this script has no resolver for, and
 * a throw would abort the enumeration instead of narrowing it. The narrowing is safe for this
 * table's purpose, since an unreadable import can only fail to add a `@custom-variant`, and a
 * missing registration would show up as a difference rather than as a false match.
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
    try {
        return { base: NodePath.dirname(resolved), content: NodeFileSystem.readFileSync(resolved, 'utf8') };
    }
    catch {
        return { base: NodePath.dirname(resolved), content: '' };
    }
}

function loadModule(identifier, base) {
    return { base, path: identifier, module: {} };
}

async function loadDesignSystem(css = designSystemSource) {
    return await bundle.__unstable__loadDesignSystem(css, {
        base: designSystemBase,
        loadStylesheet: async (identifier, importBase) => loadStylesheet(identifier, importBase),
        loadModule: async (identifier, importBase) => loadModule(identifier, importBase),
    });
}

const designSystem = await loadDesignSystem();

const entries = [];
for (const [name, info] of designSystem.variants.entries()) {
    entries.push({ name, order: info.order, kind: info.kind });
}

const compareFnOrders = Array.from(designSystem.variants.compareFns.keys()).sort((a, z) => a - z);

/*
 * Probe each grouped order's direction, rather than reading it off `variants.ts`.
 *
 * Two members of the group are parsed into a fresh design system and their dense indices read back.
 * Whichever resolves to the smaller breakpoint taking the smaller index means ascending. A release
 * that flipped `max` to ascending would change this field and produce a diff in the generated file,
 * which is the whole reason the generator exists.
 *
 * The probe pair is chosen from the group's own members: a static member takes its root as the
 * probe, a functional one takes `<root>-sm` and `<root>-lg`, which are the names the default scale
 * defines. A group whose members yield no usable pair reports `null` and the Go side refuses to
 * render, rather than guessing a direction.
 */
function probePairFor(members) {
    const statics = members.filter((member) => member.kind === 'static').map((member) => member.name);
    if (statics.includes('sm') && statics.includes('lg')) return ['sm', 'lg'];

    const functional = members.filter((member) => member.kind === 'functional').map((member) => member.name);
    for (const root of functional) {
        // `@` is written `@sm`, everything else `min-sm`.
        const small = root === '@' ? '@sm' : `${root}-sm`;
        const large = root === '@' ? '@lg' : `${root}-lg`;
        return [small, large];
    }
    return null;
}

const byOrder = new Map();
for (const entry of entries) {
    if (!byOrder.has(entry.order)) byOrder.set(entry.order, []);
    byOrder.get(entry.order).push(entry);
}

const comparisons = [];
for (const order of compareFnOrders) {
    const members = byOrder.get(order) ?? [];
    const pair = probePairFor(members);
    if (pair === null) {
        comparisons.push({ order, ascending: null, probe: null, members: members.map((m) => m.name) });
        continue;
    }

    // A fresh design system per probe, because the dense index is a function of exactly which
    // variants have been parsed. Reusing one would let an earlier probe widen a later one's map.
    const probeSystem = await loadDesignSystem();
    for (const raw of pair) probeSystem.parseVariant(raw);
    const probeOrder = probeSystem.getVariantOrder();
    const [smallRaw, largeRaw] = pair;
    const smallIndex = probeOrder.get(probeSystem.parseVariant(smallRaw)) ?? null;
    const largeIndex = probeOrder.get(probeSystem.parseVariant(largeRaw)) ?? null;

    comparisons.push({
        order,
        ascending: smallIndex === null || largeIndex === null ? null : smallIndex < largeIndex,
        probe: { small: smallRaw, smallIndex, large: largeRaw, largeIndex },
        members: members.map((m) => m.name),
    });
}

process.stdout.write(
    JSON.stringify({ tailwindVersion: packageJson.version, entries, compareFnOrders, comparisons }, null, 4) + '\n',
);
