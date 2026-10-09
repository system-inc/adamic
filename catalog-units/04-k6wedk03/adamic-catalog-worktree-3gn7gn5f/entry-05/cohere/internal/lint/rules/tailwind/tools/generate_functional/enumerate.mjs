/*
 * Capture how the engine resolves a functional candidate's value, before any handle body runs.
 *
 * The Go port of `functionalUtility` in `internal/tailwind/functionalutility.go` is checked against
 * these resolutions rather than against a reading of the TypeScript, for the reason every other
 * internal/lint/rules/tailwind/tools/generate_* tool exists: the two are different claims and they disagree where a careful reader
 * would not predict.
 *
 * # The claim this tool makes, and the weaker one it does not
 *
 * The resolved value is an intermediate. Upstream computes it inside a closure and hands it straight
 * to `desc.handle`, so nothing public returns it and there is no `resolveValue(candidate)` to call.
 * What is observable is the declarations a class compiles to, which is the resolved value after a
 * handler has already used it.
 *
 * So this measures resolution through its effect and says so rather than implying it read the
 * intermediate. For a root whose handler is the identity on its value, which most are, the resolved
 * value appears verbatim in the emitted declaration. For a root that wraps it, it appears wrapped,
 * consistently, so a port that resolves differently produces a different string in the same place.
 *
 * The strongest signal here is the rejections. A branch that must produce nothing is unambiguous:
 * `bg/50`, `w-[3px]/50` and `grow-2/3` compile to no CSS at all, and a port that resolves them to
 * something would emit a declaration where the engine emits none. Those cases carry the test.
 *
 * Usage:
 *
 *   node internal/lint/rules/tailwind/tools/generate_functional/enumerate.mjs <theme.css> [--resolve-root <dir>] > functional.json
 */

import { loadDesignSystem, parseCandidate } from '../generate_descriptors/loader.mjs';

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
 * The probe set, chosen to reach each branch rather than to be large.
 *
 * A corpus of ordinary classes exercises branch 3 thousands of times and every other branch zero
 * times, which is the shape that produces a high agreement number while testing almost nothing. Each
 * entry names the branch it is for, so coverage is readable rather than countable.
 */
const probes = [
    { className: 'grow', branch: 'no-value-default' },
    { className: 'rounded', branch: 'no-value-theme' },
    { className: 'border', branch: 'no-value-default' },
    { className: 'bg/50', branch: 'no-value-with-modifier-rejected' },
    { className: 'w-[3px]', branch: 'arbitrary' },
    { className: 'bg-[color:var(--a)]', branch: 'arbitrary-typehint' },
    { className: 'w-[3px]/50', branch: 'arbitrary-with-modifier-rejected' },
    { className: 'font-bold', branch: 'theme' },
    { className: 'bg-red-500', branch: 'theme' },
    { className: 'text-sm', branch: 'theme-nested' },
    { className: 'w-1/2', branch: 'fraction' },
    { className: 'w-1.5/2', branch: 'fraction-rejected' },
    { className: 'aspect-16/9', branch: 'fraction' },
    { className: 'p-4', branch: 'bare' },
    { className: '-m-4', branch: 'bare-negative' },
    { className: 'w-full', branch: 'static-value' },
    { className: 'w-auto', branch: 'static-value' },
    { className: 'grow-2/3', branch: 'bare-with-modifier-rejected' },
    { className: 'z-10', branch: 'bare' },
    { className: '-z-10', branch: 'bare-negative' },
];

function flatten(node) {
    const declarations = [];
    const queue = [node];
    while (queue.length > 0) {
        const current = queue.shift();
        if (!current) continue;
        if (current.kind === 'declaration') {
            declarations.push({ property: current.property, value: current.value ?? null });
            continue;
        }
        if (current.kind === 'rule' || current.kind === 'at-rule') {
            for (const child of current.nodes ?? []) queue.push(child);
        }
    }
    return declarations;
}

const entries = [];
for (const probe of probes) {
    const candidate = parseCandidate(designSystem, probe.className);
    let declarations = null;
    if (candidate) {
        try {
            const nodes = designSystem.compileAstNodes(candidate);
            if (nodes && nodes.length > 0) declarations = flatten(nodes[0].node);
        }
        catch {
            declarations = null;
        }
    }

    entries.push({
        className: probe.className,
        branch: probe.branch,
        parsed: candidate ? { kind: candidate.kind, root: candidate.root ?? null } : null,
        declarations,
    });
}

const rejected = entries.filter((entry) => entry.declarations === null).length;
if (rejected === 0) {
    process.stderr.write('enumerate.mjs: no probe was rejected, so the branches that must produce nothing are unexercised and a passing comparison would mean nothing\n');
    process.exit(4);
}

process.stdout.write(JSON.stringify({
    tailwindVersion,
    entryPoint: recordedEntryPoint,
    groundTruth: 'compileAstNodes declarations; resolution is observed through its effect rather than read directly',
    probeCount: entries.length,
    rejectedCount: rejected,
    probes: entries,
}, null, 2) + '\n');
