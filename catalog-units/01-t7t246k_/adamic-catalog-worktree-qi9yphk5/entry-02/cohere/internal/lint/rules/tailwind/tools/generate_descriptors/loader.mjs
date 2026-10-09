/*
 * Design-system loader, shared by the descriptor extractor and its harness.
 *
 * Lifted from `internal/lint/rules/tailwind/tools/generate_collapse/enumerate.mjs` lines 1-108, which is proven against two
 * repos: it walks the `@import` graph through `loadStylesheet` and takes the ESM branch of
 * Tailwind's exports map, because the `require` branch resolves to `dist/lib.js`, which does not
 * export `__unstable__loadDesignSystem` at all.
 *
 * It additionally exposes `inferDataType`, Tailwind's own type predicate, read out of the installed
 * package rather than transcribed. That matters for what this tool is testing: the claim under test
 * is that a root's reading is a function of the inferred data type, and a transcribed predicate
 * would be testing the transcription instead of the claim.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodeModule from 'node:module';
import * as NodePath from 'node:path';
import * as NodeUrl from 'node:url';
import { locate } from '../corpus.mjs';

/*
 * Resolve Tailwind's ESM entry.
 *
 * `require.resolve` takes the `require` branch of the exports map, and for Tailwind that is
 * `dist/lib.js`, which does not export `__unstable__loadDesignSystem` at all. Only the ESM build
 * does, so the package root is located through its `package.json` and the `import` condition read
 * directly.
 */
function resolveModuleUrl(specifier, directory) {
    const requireFromDirectory = NodeModule.createRequire(NodePath.join(directory, 'noop.js'));
    const packageJsonPath = requireFromDirectory.resolve(specifier + '/package.json');
    const packageJson = JSON.parse(NodeFileSystem.readFileSync(packageJsonPath, 'utf8'));
    const rootExport = packageJson.exports?.['.'] ?? packageJson.exports;
    const importEntry = typeof rootExport === 'string' ? rootExport : (rootExport?.import ?? rootExport?.default);
    const resolvedPath = importEntry
        ? NodePath.join(NodePath.dirname(packageJsonPath), importEntry)
        : requireFromDirectory.resolve(specifier);
    return NodeUrl.pathToFileURL(resolvedPath).href;
}

/*
 * A relative `@import` resolves against the importing file, and a bare one against the resolve
 * root. Keeping those two separate is what lets a design system live outside the tree that
 * installed Tailwind: `./tokens.css` still means the file beside it, while `tailwindcss` means the
 * package the caller named.
 */
function resolveStylesheet(specifier, directory, resolveRoot) {
    if (specifier.startsWith('.') || NodePath.isAbsolute(specifier)) {
        return NodePath.resolve(directory, specifier);
    }
    const requireFromDirectory = NodeModule.createRequire(NodePath.join(resolveRoot, 'noop.js'));
    if (specifier.endsWith('.css')) return requireFromDirectory.resolve(specifier);

    const packageJsonPath = requireFromDirectory.resolve(specifier + '/package.json');
    const packageJson = JSON.parse(NodeFileSystem.readFileSync(packageJsonPath, 'utf8'));
    const rootExport = packageJson.exports?.['.'] ?? packageJson.exports;
    const styleEntry = (typeof rootExport === 'object' ? rootExport?.style : undefined) ?? packageJson.style;
    if (!styleEntry) return requireFromDirectory.resolve(specifier);
    return NodePath.join(NodePath.dirname(packageJsonPath), styleEntry);
}

/*
 * `inferDataType`, recovered from the installed bundle.
 *
 * It is not on the design system and it is not a package export, so it is reached through the chunk
 * that carries it. The chunk filename is content-hashed and changes every release, so the chunk is
 * found by its export signature rather than by name: `inferDataType` is minified to `ge` and
 * re-exported as `d`, alongside nine siblings. Probing candidate chunks for a function that types
 * `url(a.png)` as `image` and `3px` as `length` identifies it without depending on the hash.
 *
 * If a future Tailwind reshapes the chunk, this returns null and the caller reports the descriptor
 * model as untested rather than silently falling back to a hand-rolled predicate, which would test
 * the fallback instead of the engine.
 */
async function loadInferDataType(tailwindModuleUrl) {
    const distDirectory = NodePath.dirname(NodeUrl.fileURLToPath(tailwindModuleUrl));
    let chunkNames;
    try {
        chunkNames = NodeFileSystem.readdirSync(distDirectory).filter((name) => name.endsWith('.mjs'));
    }
    catch {
        return null;
    }

    for (const chunkName of chunkNames) {
        let chunkModule;
        try {
            chunkModule = await import(NodeUrl.pathToFileURL(NodePath.join(distDirectory, chunkName)).href);
        }
        catch {
            continue;
        }
        for (const exported of Object.values(chunkModule)) {
            if (typeof exported !== 'function' || exported.length !== 2) continue;
            try {
                const probeTypes = ['image', 'length', 'percentage', 'color'];
                if (exported('url(a.png)', probeTypes) !== 'image') continue;
                if (exported('3px', probeTypes) !== 'length') continue;
                if (exported('50%', probeTypes) !== 'percentage') continue;
                if (exported('red', probeTypes) !== 'color') continue;
                if (exported('var(--a)', probeTypes) !== null) continue;
                return exported;
            }
            catch {
                continue;
            }
        }
    }
    return null;
}

/*
 * `resolveRootArgument` is where bare specifiers resolve from, which is a different question from
 * where the CSS lives. For a repository they have the same answer: its stylesheet sits inside the
 * tree that installed Tailwind, so anchoring resolution on the entry point's own directory worked
 * and the distinction stayed invisible.
 *
 * It stops working for exactly the file that matters most. A design system written to share no
 * `@theme` with any repository is checked in under `tools/` with no `node_modules` above it, so
 * `@import 'tailwindcss'` cannot resolve and the file is ungenerable, which is the state
 * `testdata/independent_theme.css` was in: the only stylesheet capable of showing this tool
 * describing bare Tailwind was the one stylesheet this tool could not read.
 *
 * The install is borrowed rather than vendored, matching `generate_collapse`. A second Tailwind
 * checked in beside a probe theme would be a version to keep in step, and nothing of the lending
 * repository's design system reaches the probe, because only the CSS named as the entry point is
 * loaded.
 *
 * Defaults to the entry point's directory, so every repository invocation is unchanged.
 *
 * `entryPointArgument` may be a corpus spelling such as `ahra:app/_theme/styles/theme.css` (see
 * `../corpus.mjs`). `entryPoint` is then the file it names, and `recordedEntryPoint` the spelling,
 * which is what a fixture records so it reads the same on every machine.
 */
export async function loadDesignSystem(entryPointArgument, resolveRootArgument, entrySourceOverride) {
    const entryPointLocated = locate(entryPointArgument);
    const entryPointPath = entryPointLocated.path;
    const entryDirectory = NodePath.dirname(entryPointPath);
    const resolveRoot = NodePath.resolve(resolveRootArgument ?? entryDirectory);
    const tailwindModuleUrl = resolveModuleUrl('tailwindcss', resolveRoot);
    const tailwindModule = await import(tailwindModuleUrl);

    /*
     * `entrySourceOverride` supplies the entry stylesheet inline instead of reading it. The bare
     * baseline in `surveyOwnContributions` is one literal `@import`, and writing it to disk to read
     * it straight back would leave a file in a repository this tool is only borrowing an install
     * from. Its path is still notional so `base` anchors correctly.
     */
    const designSystem = await tailwindModule.__unstable__loadDesignSystem(
        entrySourceOverride ?? NodeFileSystem.readFileSync(entryPointPath, 'utf8'),
        {
            base: entryDirectory,
            loadModule: async function (specifier, base, resourceType) {
                try {
                    /*
                     * A relative `@plugin` is resolved against the importing file, as before; a bare
                     * one falls back to the resolve root, so a borrowed install can supply it. No
                     * stylesheet in the corpus uses `@plugin` or `@config`, so this path is
                     * currently unexercised and the fallback is the only behaviour that changes.
                     */
                    const moduleBase = specifier.startsWith('.') || NodePath.isAbsolute(specifier) ? base : resolveRoot;
                    const requireFromBase = NodeModule.createRequire(NodePath.join(moduleBase, 'noop.js'));
                    const resolved = requireFromBase.resolve(specifier);
                    const loaded = await import(NodeUrl.pathToFileURL(resolved).href);
                    return { base: NodePath.dirname(resolved), module: loaded.default ?? loaded };
                }
                catch {
                    return { base, module: resourceType === 'config' ? {} : () => {} };
                }
            },
            loadStylesheet: async function (specifier, base) {
                try {
                    const resolved = resolveStylesheet(specifier, base, resolveRoot);
                    return { base: NodePath.dirname(resolved), content: NodeFileSystem.readFileSync(resolved, 'utf8') };
                }
                catch {
                    return { base: '', content: '' };
                }
            },
        },
    );

    const requireFromResolveRoot = NodeModule.createRequire(NodePath.join(resolveRoot, 'noop.js'));
    const tailwindVersion = JSON.parse(
        NodeFileSystem.readFileSync(requireFromResolveRoot.resolve('tailwindcss/package.json'), 'utf8'),
    ).version;

    return {
        designSystem,
        tailwindVersion,
        entryPoint: entryPointPath,
        recordedEntryPoint: entryPointLocated.recorded,
        resolveRoot,
        inferDataType: await loadInferDataType(tailwindModuleUrl),
    };
}

/*
 * What this design system registers that a bare Tailwind install does not.
 *
 * The volume assertions cannot answer this on their own, and that is the gap they were found to
 * have. Tailwind registers its built-in utility roots from JavaScript rather than from CSS, so a
 * stylesheet importing nothing but the framework still produces 23,286 registry classes, 1,154
 * utility roots and 419 theme keys, and every floor written to catch a loader that returned an
 * empty system is cleared comfortably. Measured, not assumed: `@import "tailwindcss" source(none);`
 * passed all six assertions before this existed.
 *
 * The delta is what discriminates, and it is counted rather than named. A hardcoded canary
 * (`fade-in` exists on ahra, `brand-hover` on connected) is the same baked-in assumption the check
 * is meant to catch, one repository's tokens standing in for a claim about any repository: it
 * cannot travel, and it goes quiet the day someone renames the utility. Comparing against a live
 * bare install asks the question directly, in whatever repository it runs, and needs no maintenance
 * when the framework's own population moves.
 *
 * The baseline is built through the same borrowed install as the system under test, so the two
 * populations come from one engine. Differencing against a different Tailwind version would
 * attribute that version's own additions to the repository, which is the error this is here to
 * prevent, pointing the other way.
 */
export async function surveyOwnContributions(designSystem, resolveRoot) {
    const baseline = await loadDesignSystem(
        NodePath.join(resolveRoot, '__bare_tailwind_probe__.css'),
        resolveRoot,
        '@import "tailwindcss" source(none);',
    );

    function rootsOf(system) {
        const roots = new Set();
        for (const entry of system.getClassList()) {
            const candidate = parseCandidate(system, entry[0]);
            if (candidate?.root) roots.add(candidate.root);
        }
        return roots;
    }

    function themeKeysOf(system) {
        return new Set(Array.from(system.theme.entries(), ([key]) => key));
    }

    const baselineRoots = rootsOf(baseline.designSystem);
    const baselineThemeKeys = themeKeysOf(baseline.designSystem);
    const ownRoots = Array.from(rootsOf(designSystem)).filter((root) => !baselineRoots.has(root)).sort();
    const ownThemeKeys = Array.from(themeKeysOf(designSystem)).filter((key) => !baselineThemeKeys.has(key)).sort();

    return {
        baselineTailwindVersion: baseline.tailwindVersion,
        baselineRegistryClasses: baseline.designSystem.getClassList().length,
        baselineUtilityRoots: baselineRoots.size,
        baselineThemeKeys: baselineThemeKeys.size,
        ownUtilityRootCount: ownRoots.length,
        ownThemeKeyCount: ownThemeKeys.length,
        ownUtilityRoots: ownRoots,
        ownThemeKeys: ownThemeKeys,
    };
}

/*
 * The reading of a class, as the engine computes it.
 *
 * `compileAstNodes` is an internal API and upstream may rename it. It is used here rather than the
 * public `getClassOrder` because the question is diagnostic: `getClassOrder` returns a position
 * within the queried set, which answers "did these two classes agree" but not "what did this class
 * declare". `{order, count}` is the thing the Go descriptor table has to reproduce, so it is the
 * thing that has to be measured. `getClassOrder` is used separately as the end-to-end control.
 *
 * `nodes[0]` is deliberate. A candidate can parse several ways (`border-b` parses two), and
 * `getClassOrder` takes the position of the first (`sort.ts:22`). Taking the first here matches it.
 *
 * A null return means the class produced no reading at all. That is not agreement and callers must
 * not count it as one: in the ahra corpus 10 of 1,208 distinct classes return null, and a harness
 * that scored them as passes would be blind on exactly the classes most likely to break.
 */
export function readingOf(designSystem, className) {
    let candidate;
    try {
        candidate = Array.from(designSystem.parseCandidate?.(className) ?? [])[0];
    }
    catch {
        return null;
    }
    if (!candidate) return null;

    let nodes;
    try {
        nodes = designSystem.compileAstNodes(candidate);
    }
    catch {
        return null;
    }
    if (!nodes || nodes.length === 0) return null;

    const propertySort = nodes[0]?.propertySort;
    if (!propertySort) return null;
    return { order: propertySort.order ?? [], count: propertySort.count ?? 0 };
}

export function parseCandidate(designSystem, className) {
    try {
        return Array.from(designSystem.parseCandidate?.(className) ?? [])[0] ?? null;
    }
    catch {
        return null;
    }
}

export function readingKey(reading) {
    if (reading === null) return 'null';
    return '[' + reading.order.join(',') + ']#' + reading.count;
}
