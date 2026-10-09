/*
 * Ask the real Tailwind engine for compiled utility trees and the reading it takes from each, and
 * emit both as JSON.
 *
 * The Go port of the AST and PropertySort is checked against these answers rather than against a
 * reading of the TypeScript source. The distinction is not academic here. `getPropertySort` is a
 * breadth-first queue, not a call to `walk`, and it descends only into `rule` and `at-rule`; both
 * facts are invisible in a casual read and both change answers. Every expectation in this corpus is
 * a tree the shipped engine built and a reading the shipped engine took from it.
 *
 * The engine is reached through `__unstable__loadDesignSystem`, which is a real public entry point
 * of the package, so no bundle spelunking is needed. `compileAstNodes` hands back `{node,
 * propertySort}` per compiled rule: the tree and its reading, together, from the same call.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodePath from 'node:path';

const [, , packageRootArgument] = process.argv;
if (!packageRootArgument) {
    process.stderr.write('usage: enumerate.mjs <tailwindcss package root>\n');
    process.exit(2);
}

const packageRoot = NodePath.resolve(packageRootArgument);
const packageJson = JSON.parse(NodeFileSystem.readFileSync(NodePath.join(packageRoot, 'package.json'), 'utf8'));

const { __unstable__loadDesignSystem } = await import(NodePath.join(packageRoot, 'dist', 'lib.mjs'));

/*
 * Load with the package's own index.css so the real theme is present. Without it most utilities
 * resolve to nothing and the corpus would be a few dozen structural shapes rather than the whole
 * class list.
 */
const design = await __unstable__loadDesignSystem('@import "tailwindcss";', {
    base: packageRoot,
    loadStylesheet: async (id, base) => {
        let file = id.replace(/^tailwindcss\/?/, '') || 'index.css';
        if (!file.endsWith('.css')) file += '.css';
        const path = NodePath.join(packageRoot, file);
        return { base: NodePath.dirname(path), content: NodeFileSystem.readFileSync(path, 'utf8'), path };
    },
});

/*
 * Serialize a node to exactly the fields the Go port models. `src`/`dst` source locations are
 * dropped because the port does not carry them; everything else is kept verbatim, including the
 * distinction between a declaration whose value is `undefined` and one whose value is ''.
 */
function serializeNode(node) {
    switch (node.kind) {
        case 'rule':
            return { kind: 'rule', selector: node.selector, nodes: node.nodes.map(serializeNode) };
        case 'at-rule':
            return { kind: 'at-rule', name: node.name, params: node.params, nodes: node.nodes.map(serializeNode) };
        case 'declaration':
            return {
                kind: 'declaration',
                property: node.property,
                value: node.value === undefined || node.value === null ? null : node.value,
                valuePresent: !(node.value === undefined || node.value === null),
                important: Boolean(node.important),
            };
        case 'comment':
            return { kind: 'comment', value: node.value };
        case 'context':
            return {
                kind: 'context',
                context: Object.fromEntries(Object.entries(node.context).map(([k, v]) => [k, String(v)])),
                nodes: node.nodes.map(serializeNode),
            };
        case 'at-root':
            return { kind: 'at-root', nodes: node.nodes.map(serializeNode) };
        default:
            throw new Error(`Unknown node kind: ${node.kind}`);
    }
}

/*
 * The full class list is 23k candidates and most are the same shape with a different value, which
 * makes for a large fixture that proves little beyond its first few hundred rows. Bucket by
 * structural signature instead and keep a bounded number of each, so every distinct tree shape the
 * engine can build is represented and near-duplicates are not.
 */
function signature(node) {
    switch (node.kind) {
        case 'declaration':
            return `d:${node.property}`;
        case 'comment':
            return 'c';
        default:
            return `${node.kind}(${(node.nodes ?? []).map(signature).join(',')})`;
    }
}

const perSignature = new Map();
const perSignatureLimit = 3;
let compiledCount = 0;

const classNames = design.getClassList().map(([name]) => name);

/*
 * Hand-chosen shapes alongside the generated list. These carry the behaviors the class list happens
 * to exercise thinly or not at all: arbitrary values, variants (which must not change the reading),
 * important, and the bare `--tw-sort` latch.
 */
const extraClassNames = [
    'shadow-[--x]',
    'shadow-[0_0_1px_red]',
    'flex',
    'p-4',
    '-space-x-4',
    'space-y-2',
    'divide-x-2',
    'bg-linear-30',
    '-bg-linear-30',
    'bg-conic-180',
    'sm:hover:underline',
    'dark:md:flex',
    'p-4!',
    'ring-2',
    'ring-offset-2',
    'transition',
    'animate-spin',
    'grid-cols-[repeat(3,minmax(0,1fr))]',
    'w-[calc(100%-var(--x))]',
    'text-(--my-color)',
    'bg-red-500/50',
    'content-["hello"]',
    'not-hover:flex',
    'group-hover:opacity-50',
    '@md:flex',
    /*
     * Variants whose expansion duplicates the body. These are the cases that distinguish the
     * pre-variant list the reading is taken from and the post-variant tree the engine emits; without
     * them the corpus proves variant-invariance only on variants that happen to preserve the shape.
     */
    'not-focus:underline',
    'not-motion-safe:flex',
    'not-dark:flex',
    'not-print:block',
    'not-supports-grid:flex',
    'hover:flex',
    'focus:underline',
    'dark:flex',
    'sm:flex',
    'motion-safe:transition',
    'peer-checked:block',
    'has-hover:flex',
    'in-hover:flex',
    'aria-expanded:block',
    'data-open:flex',
    'nth-3:flex',
    'first:mt-0',
    'last:mb-0',
    'odd:bg-red-500',
    'open:block',
    'sm:hover:not-focus:underline',
];

const cases = [];

for (const name of [...extraClassNames, ...classNames]) {
    let candidates;
    try {
        candidates = design.parseCandidate(name);
    } catch {
        continue;
    }
    if (!candidates || candidates.length === 0) continue;

    for (const candidate of candidates) {
        let compiled;
        try {
            compiled = design.compileAstNodes(candidate);
        } catch {
            continue;
        }
        if (!compiled || compiled.length === 0) continue;

        /*
         * `getPropertySort` is seeded with the utility's declaration list *before* the wrapping
         * `.selector` rule and *before* any variant is applied. The tree `compileAstNodes` returns
         * is the post-variant one, and for a duplicating variant like `not-hover:` the two differ:
         * the returned tree holds two copies of the body while the reading counts one. So the
         * candidate is recompiled with its variants stripped to recover the exact list the engine
         * read, and both trees are recorded.
         *
         * The two readings are asserted equal below. That is the measurement behind the claim that
         * variants never change a candidate's reading, which would otherwise be an inference from
         * the call site.
         */
        let bareCompiled = null;
        if (candidate.variants && candidate.variants.length > 0) {
            try {
                bareCompiled = design.compileAstNodes({ ...candidate, variants: [] });
            } catch {
                bareCompiled = null;
            }
        }

        for (const [index, entry] of compiled.entries()) {
            compiledCount++;
            const readNode = serializeNode((bareCompiled?.[index] ?? entry).node);
            const emittedNode = serializeNode(entry.node);
            const key = signature(readNode);
            const seen = perSignature.get(key) ?? 0;
            const forced = extraClassNames.includes(name);
            if (!forced && seen >= perSignatureLimit) continue;
            perSignature.set(key, seen + 1);

            const bareSort = bareCompiled?.[index]?.propertySort ?? entry.propertySort;
            if (bareSort.count !== entry.propertySort.count || bareSort.order.join(',') !== entry.propertySort.order.join(',')) {
                throw new Error(
                    `${name}: reading changed when variants were stripped ` +
                        `(${JSON.stringify(entry.propertySort)} vs ${JSON.stringify(bareSort)}); ` +
                        'the assumption that variants never affect the reading is wrong',
                );
            }

            cases.push({
                className: name,
                signature: key,
                /*
                 * readNode is what getPropertySort was seeded with: the pre-variant tree. It is the
                 * input the Go port is asked to reproduce a reading from.
                 */
                node: readNode,
                /*
                 * emittedNode is the post-variant tree the engine actually returns, recorded only
                 * where it can differ from the tree the reading was taken from. For a candidate with
                 * no variants the two are the same object and storing it twice would double the
                 * fixture for no added claim.
                 */
                emittedNode: candidate.variants && candidate.variants.length > 0 ? emittedNode : null,
                hasVariants: Boolean(candidate.variants && candidate.variants.length > 0),
                /*
                 * The reading the engine itself took, from the same call that produced the tree.
                 * This is the value the Go port must reproduce.
                 */
                propertySort: { order: entry.propertySort.order, count: entry.propertySort.count },
            });
        }
    }
}

/*
 * Record the traversal order the engine's queue produces, for a handful of trees where a
 * depth-first walk would visit differently. Without these the suite would pass with a depth-first
 * PropertySort, because a reordering usually cancels out in a sorted set.
 */
function breadthFirstProperties(nodes) {
    const out = [];
    const q = nodes.slice();
    while (q.length > 0) {
        const n = q.shift();
        if (n.kind === 'declaration') {
            if (n.valuePresent) out.push(n.property);
        } else if (n.kind === 'rule' || n.kind === 'at-rule') {
            for (const child of n.nodes) q.push(child);
        }
    }
    return out;
}

function depthFirstProperties(nodes) {
    const out = [];
    const stack = nodes.slice().reverse();
    while (stack.length > 0) {
        const n = stack.pop();
        if (n.kind === 'declaration') {
            if (n.valuePresent) out.push(n.property);
        } else if (n.kind === 'rule' || n.kind === 'at-rule') {
            for (let i = n.nodes.length - 1; i >= 0; i--) stack.push(n.nodes[i]);
        }
    }
    return out;
}

const traversalCases = [];
for (const testCase of cases) {
    const bfs = breadthFirstProperties([testCase.node]);
    const dfs = depthFirstProperties([testCase.node]);
    if (bfs.join('|') !== dfs.join('|')) {
        traversalCases.push({ className: testCase.className, node: testCase.node, breadthFirst: bfs, depthFirst: dfs });
    }
}

/*
 * Count how much of the real class list diverges, over the whole list rather than the sampled
 * corpus, so the number in the Go doc comment is a measurement that regenerates rather than a
 * remembered one.
 */
let fullListCompiled = 0;
let fullListDivergent = 0;
let fullListAtRoot = 0;
for (const name of classNames) {
    let candidates;
    try {
        candidates = design.parseCandidate(name);
    } catch {
        continue;
    }
    if (!candidates) continue;
    for (const candidate of candidates) {
        let compiled;
        try {
            compiled = design.compileAstNodes(candidate);
        } catch {
            continue;
        }
        for (const entry of compiled) {
            fullListCompiled++;
            const serialized = serializeNode(entry.node);
            if (JSON.stringify(serialized).includes('"at-root"')) fullListAtRoot++;
            const bfs = breadthFirstProperties([serialized]).join('|');
            const dfs = depthFirstProperties([serialized]).join('|');
            if (bfs !== dfs) fullListDivergent++;
        }
    }
}

process.stdout.write(
    JSON.stringify(
        {
            tailwindVersion: packageJson.version,
            classListSize: classNames.length,
            compiledCount,
            distinctSignatures: perSignature.size,
            fullList: {
                compiled: fullListCompiled,
                breadthFirstDiffersFromDepthFirst: fullListDivergent,
                withAtRootSubtree: fullListAtRoot,
            },
            cases,
            traversalCases,
        },
        null,
        2,
    ) + '\n',
);
