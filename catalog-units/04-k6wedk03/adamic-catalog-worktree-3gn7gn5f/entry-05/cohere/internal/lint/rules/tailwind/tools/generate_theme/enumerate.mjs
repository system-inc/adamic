/*
 * Ask the real Tailwind 4.3.3 engine what theme it resolved for a corpus of stylesheets, and emit
 * both the resolved theme and the answers to every query the Go port must reproduce.
 *
 * The Go port of `theme.ts` is checked against these answers rather than against a reading of the
 * TypeScript source, for the reason generate_data_type and generate_css_parser exist: the two
 * are different claims and they disagree in ways a careful reader would not predict. The one that
 * decides this component is the ignored-key map. Reading the source suggests `--font` names a
 * namespace holding every key starting `--font-`; asking the engine says `--font-weight-*` and
 * `--font-size-*` are excluded from it, so this repository's `--color` namespace holds 567 keys
 * while `keysInNamespaces(['--color'])` returns 558, and the gap is invisible unless a fixture
 * measures both.
 *
 * # Why this tool asks the design system rather than the Theme class
 *
 * `Theme` is not an export of the published package. It is reachable, though, through a documented
 * export: `__unstable__loadDesignSystem` returns a design system whose `theme` property is a live
 * `Theme` instance built from the stylesheet's whole `@import` graph. That is a better ground truth
 * than a re-exported constructor would be, because it is the theme the shipped engine actually
 * assembled from the repository in front of it, `@import` resolution, `@theme` option parsing,
 * `initial` deletions and all.
 *
 * So this tool never needs the bundle-patching that generate_css_parser needs. The public route
 * exists here, and it exposes exactly the object under test.
 *
 * # The two corpora, and why both
 *
 * `real` cases load an actual repository's `theme.css` and its whole import graph. That is the exit
 * criterion for this component: a resolved namespace map diffed against the engine's own theme, on
 * two repositories, because one design system is not a population and the entire argument for the
 * port is that two repositories on the same Tailwind differ.
 *
 * `synthetic` cases are small stylesheets that reach the branches a real theme never does:
 * `initial` deletion, `--*: initial` wholesale clearing, namespace clearing, `@theme default`
 * losing to a non-default value, `inline` and `reference` option handling, prefixes, and the
 * dot-to-underscore fallback in key resolution. A real theme exercises perhaps half the branches in
 * `theme.ts`; the other half is where a port breaks silently.
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
 * Resolve `@import` the way a bundler would. `tailwindcss` and `tailwindcss/theme` name files
 * inside the package root; everything else is relative to the importing file. The engine appends no
 * extension itself, so a bare specifier gets `.css`.
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
        path: resolved,
    };
}

/*
 * `@config` and `@plugin` load JavaScript. cohere never runs either, and neither contributes theme
 * entries through this path, so the module is stubbed rather than executed. A real `@config` that
 * defined theme values would be a gap; the two repositories in the corpus use `@config` only for
 * content globs.
 */
function loadModule(identifier, base) {
    return { base, path: identifier, module: {} };
}

async function loadTheme(entryCss, entryBase) {
    const designSystem = await bundle.__unstable__loadDesignSystem(entryCss, {
        base: entryBase,
        loadStylesheet: async (identifier, base) => loadStylesheet(identifier, base),
        loadModule: async (identifier, base) => loadModule(identifier, base),
    });
    return designSystem.theme;
}

/*
 * Every namespace the engine's own theme actually contains, derived from the keys rather than from
 * a hardcoded list, so a repository that invents a namespace is covered without editing this file.
 *
 * A key like `--color-red-500` could be read as the namespace `--color` or `--color-red`, and both
 * are legitimate: `--inset-shadow-sm` really does live in the `--inset-shadow` namespace, not in
 * `--inset`. So every prefix of every key is offered as a candidate namespace, which over-covers on
 * purpose. The port has to agree on all of them, including the ones that resolve to nothing.
 */
function candidateNamespaces(theme) {
    const namespaces = new Set();
    for (const [key] of theme.entries()) {
        const subVariableIndex = key.indexOf('--', 2);
        const withoutSubVariable = subVariableIndex === -1 ? key : key.slice(0, subVariableIndex);
        const parts = withoutSubVariable.slice(2).split('-');
        for (let index = 1; index <= parts.length; index++) {
            namespaces.add(`--${parts.slice(0, index).join('-')}`);
        }
    }
    return [...namespaces].sort();
}

/*
 * Serialize a theme into everything the port must reproduce.
 *
 * `entries` carries insertion order, not a sorted key list, because insertion order is observable:
 * `keysInNamespaces` and `namespace()` both return values in it, and a Go port backed by a plain
 * map would return a different order on every run while passing any test that sorted first.
 */
function serializeTheme(theme) {
    const entries = [...theme.entries()].map(([key, value]) => ({
        key,
        value: value.value,
        options: value.options,
    }));

    const namespaces = candidateNamespaces(theme);

    const namespaceResults = namespaces.map((namespace) => ({
        namespace,
        // `namespace()` keys `null` for the exact match; JSON has no distinguished null map key, so
        // it is carried as an explicit flag alongside.
        entries: [...theme.namespace(namespace)].map(([key, value]) => ({
            key: key === null ? '' : key,
            keyIsNull: key === null,
            value,
        })),
        keysInNamespace: theme.keysInNamespaces([namespace]),
    }));

    /*
     * Resolution queries. Every key in the theme is asked back under the namespace it came from,
     * which is the query the `@utility` evaluator will make, plus deliberate misses.
     */
    const resolutions = [];
    const seen = new Set();
    function ask(candidateValue, themeKeys) {
        const signature = `${candidateValue} ${themeKeys.join('')}`;
        if (seen.has(signature)) return;
        seen.add(signature);
        resolutions.push({
            candidateValue: candidateValue === null ? '' : candidateValue,
            candidateValueIsNull: candidateValue === null,
            themeKeys,
            resolve: theme.resolve(candidateValue, themeKeys),
            resolveValue: theme.resolveValue(candidateValue, themeKeys),
        });
    }

    for (const { key } of entries) {
        const subVariableIndex = key.indexOf('--', 2);
        const withoutSubVariable = subVariableIndex === -1 ? key : key.slice(0, subVariableIndex);
        const parts = withoutSubVariable.slice(2).split('-');
        // Ask under every split point, so `--color-red-500` is asked as `--color` + `red-500` and
        // as `--color-red` + `500`, and the ignored-key map is exercised from both sides.
        for (let index = 1; index < parts.length; index++) {
            ask(parts.slice(index).join('-'), [`--${parts.slice(0, index).join('-')}`]);
        }
        ask(null, [withoutSubVariable]);
    }

    // Deliberate misses and multi-key fallback, the shapes a real theme never produces.
    ask('does-not-exist', ['--color']);
    ask('red-500', ['--nope', '--color']);
    ask('red-500', ['--color', '--nope']);
    ask(null, ['--nope']);
    ask('1.5', ['--spacing']);
    ask('weight-bold', ['--font']);
    ask('size-lg', ['--font']);
    ask('shadow-sm', ['--inset']);
    ask('color', ['--text']);
    ask('', ['--color']);

    /*
     * `resolveWith` with the nested keys the engine itself uses. This is the `-*--nested` suffix
     * form the task names, and it is what makes `text-sm` emit a line-height alongside a size.
     */
    const nestedKeyGroups = [
        ['--text', ['--line-height', '--letter-spacing', '--font-weight']],
        ['--font', ['--font-feature-settings', '--font-variation-settings']],
        ['--shadow', ['--color']],
    ];
    const resolveWithResults = [];
    for (const [namespace, nestedKeys] of nestedKeyGroups) {
        for (const candidateValue of theme.keysInNamespaces([namespace])) {
            const result = theme.resolveWith(candidateValue, [namespace], nestedKeys);
            resolveWithResults.push({
                candidateValue,
                themeKeys: [namespace],
                nestedKeys,
                // `null` when the key does not resolve; otherwise `[value, extra]`.
                value: result === null ? null : result[0],
                extra: result === null ? null : result[1],
            });
        }
    }

    /*
     * `get`, `hasDefault` and `getOptions` over every key, plus a miss. These are the queries that
     * read the options bitfield, and the bitfield is what decides whether `resolve` returns a
     * `var()` or the literal value.
     */
    const optionQueries = entries.map(({ key }) => ({
        key,
        get: theme.get([key]),
        hasDefault: theme.hasDefault(key),
        options: theme.getOptions(key),
    }));
    optionQueries.push({
        key: '--does-not-exist',
        get: theme.get(['--does-not-exist']),
        hasDefault: theme.hasDefault('--does-not-exist'),
        options: theme.getOptions('--does-not-exist'),
    });

    return {
        prefix: theme.prefix,
        size: theme.size,
        entries,
        namespaceResults,
        resolutions,
        resolveWithResults,
        optionQueries,
    };
}

/*
 * The synthetic corpus: small stylesheets reaching the branches a real theme never does.
 *
 * Each is a complete stylesheet passed to the engine as the entry file, so the `@theme` handling in
 * `index.ts` runs exactly as it does for a repository, including option parsing and prefixes.
 */
const syntheticStylesheets = [
    ['empty theme', '@theme {}'],
    ['single value', '@theme { --color-a: red; }'],
    ['two values', '@theme { --color-a: red; --color-b: blue; }'],
    ['insertion order preserved', '@theme { --color-z: red; --color-a: blue; --color-m: green; }'],

    // `initial` deletes rather than storing.
    ['initial deletes', '@theme { --color-a: red; --color-a: initial; }'],
    ['initial on absent key', '@theme { --color-a: initial; --color-b: blue; }'],
    ['redefinition keeps position', '@theme { --color-a: red; --color-b: blue; --color-a: green; }'],

    // Namespace clearing. `--color-*: initial` clears the namespace; `--*: initial` clears all.
    ['namespace clear', '@theme { --color-a: red; --color-*: initial; --color-b: blue; }'],
    ['clear everything', '@theme { --color-a: red; --spacing: 1px; --*: initial; --color-b: blue; }'],
    ['namespace clear is prefix based', '@theme { --color-a: red; --colorful-b: blue; --color-*: initial; }'],
    ['clear ignored subnamespace survives', '@theme { --font-a: x; --font-weight-b: y; --font-*: initial; }'],
    ['clear inset keeps inset-shadow', '@theme { --inset-a: x; --inset-shadow-b: y; --inset-*: initial; }'],

    // `@theme default` yields to an existing non-default value.
    ['default does not override', '@theme { --color-a: red; }\n@theme default { --color-a: blue; }'],
    ['default overrides default', '@theme default { --color-a: red; }\n@theme default { --color-a: blue; }'],
    ['non-default overrides default', '@theme default { --color-a: red; }\n@theme { --color-a: blue; }'],

    // Option bits, read back through getOptions and observed through resolve.
    ['inline option', '@theme inline { --color-a: red; }'],
    ['reference option', '@theme reference { --color-a: red; }'],
    ['static option', '@theme static { --color-a: red; }'],
    ['inline and reference', '@theme inline reference { --color-a: red; }'],
    ['default and inline', '@theme default inline { --color-a: red; }'],

    // Prefix changes every emitted variable name and unprefixes on the way in.
    ['prefix', '@theme prefix(tw) { --color-a: red; }'],
    ['prefix with nested', '@theme prefix(tw) { --text-sm: 1rem; --text-sm--line-height: 2; }'],

    // The `-*--nested` suffix form, standalone.
    ['nested sub variable', '@theme { --text-sm: 1rem; --text-sm--line-height: 1.5; }'],
    ['nested without parent', '@theme { --text-sm--line-height: 1.5; }'],
    ['double nested', '@theme { --text-sm: 1rem; --text-sm--line-height--x: 2; }'],

    // The dot-to-underscore fallback in key resolution.
    ['underscore key for dotted value', '@theme { --spacing-1_5: 0.375rem; }'],
    ['dotted key literally', '@theme { --spacing-1\\.5: 0.375rem; }'],

    // Ignored-key map, every entry of it.
    ['font ignores weight and size', '@theme { --font-a: x; --font-weight-b: y; --font-size-c: z; }'],
    ['inset ignores shadow and ring', '@theme { --inset-a: x; --inset-shadow-b: y; --inset-ring-c: z; }'],
    [
        'text ignores its six',
        '@theme { --text-a: x; --text-color-b: y; --text-decoration-color-c: z; --text-decoration-thickness-d: w; --text-indent-e: v; --text-shadow-f: u; --text-underline-offset-g: t; }',
    ],
    [
        'grid column ignores start and end',
        '@theme { --grid-column-a: x; --grid-column-start-b: y; --grid-column-end-c: z; }',
    ],
    ['grid row ignores start and end', '@theme { --grid-row-a: x; --grid-row-start-b: y; --grid-row-end-c: z; }'],
    // The ignored key is matched by exact equality *or* by `startsWith(key + '-')`, so a key that
    // shares a prefix without the dash boundary is not ignored.
    ['ignored prefix needs the dash', '@theme { --font-weightless: x; --font-weight: y; }'],

    // Empty and odd values.
    ['empty value', '@theme { --color-a: ; }'],
    ['value with newline', '@theme { --font-sans: a,\n    b; }'],
    ['escaped key', '@theme { --color-a\\/b: red; }'],
];

const cases = [];

for (const [name, content] of syntheticStylesheets) {
    const theme = await loadTheme(content, packageRoot);
    cases.push({ name, source: 'synthetic', input: content, theme: serializeTheme(theme) });
}

const realFiles = corpusPathArgument
    ? JSON.parse(NodeFileSystem.readFileSync(NodePath.resolve(corpusPathArgument), 'utf8'))
    : [];

for (const entry of realFiles) {
    const entryLocated = locate(entry.path);
    const content = NodeFileSystem.readFileSync(entryLocated.path, 'utf8');
    const theme = await loadTheme(content, NodePath.dirname(entryLocated.path));
    cases.push({
        name: entry.name,
        source: 'repository',
        // The real cases carry the entry path rather than the content: the Go side resolves the
        // same `@import` graph from disk, which is the thing under test. A corpus spelling is
        // recorded as given, so the fixture reads the same on every machine.
        entryPath: entryLocated.recorded,
        theme: serializeTheme(theme),
    });
}

process.stdout.write(
    JSON.stringify(
        {
            tailwindVersion: packageJson.version,
            syntheticCaseCount: syntheticStylesheets.length,
            repositoryCaseCount: realFiles.length,
            cases,
        },
        null,
        2,
    ) + '\n',
);
