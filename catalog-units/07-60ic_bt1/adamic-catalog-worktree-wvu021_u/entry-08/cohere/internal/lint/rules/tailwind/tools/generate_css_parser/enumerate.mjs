/*
 * Ask the real Tailwind 4.3.3 parser what it built for a corpus of stylesheets, and emit the trees
 * as JSON.
 *
 * The Go port of `css-parser.ts` is checked against these trees rather than against a reading of
 * the TypeScript source, for the same reason generate_data_type and generate_syntax_tree exist: the
 * two are different claims and they disagree in ways a careful reader would not predict. The one
 * that decides this component is the custom-property branch. Reading the source suggests a `--foo`
 * declaration comes back whitespace-normalized like every other declaration; asking the engine says
 * its value keeps the literal newlines and indentation of the source file, because that branch
 * slices the input directly and never routes those bytes through the buffer.
 *
 * `parse` is not an export of the published package. The bundle's only exports are `Features`,
 * `Polyfills`, `__unstable__loadDesignSystem`, `compile` and `compileAst`, and the public routes
 * were tried first and rejected: `compile().build()` returns its input untouched when no Tailwind
 * feature is engaged, so round-tripping a stylesheet through it proves nothing about parsing, and
 * `__unstable__loadDesignSystem` exposes resolved theme values rather than the tree.
 *
 * So the bundle's own source text is re-exported into a temporary module inside `dist/` (inside,
 * because its relative chunk imports must still resolve) and the parser is located by behavior. The
 * minified name is `Te` at 4.3.3 and is not relied on: names, chunk filenames and export lists all
 * change between releases, so hardcoding any of them would break as a missing file rather than as a
 * stale assumption.
 *
 * The identity check has to be specific. Two functions in the bundle turn `.a{color:red}` into a
 * one-element array holding a rule whose selector is `.a`, and only one of them is the parser: the
 * other is a wrapper that yields `{kind:'rule', selector:'color:red', nodes:[null]}` for the body.
 * Requiring that a plain declaration parse as a `declaration`, and that a custom property keep a raw
 * newline, separates them. The check fails loudly rather than silently testing the wrong function.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodePath from 'node:path';
import * as NodeUrl from 'node:url';

const [, , packageRootArgument, corpusPathArgument] = process.argv;
if (!packageRootArgument) {
    process.stderr.write('usage: enumerate.mjs <tailwindcss package root> [corpus.json]\n');
    process.exit(2);
}

const packageRoot = NodePath.resolve(packageRootArgument);
const packageJson = JSON.parse(NodeFileSystem.readFileSync(NodePath.join(packageRoot, 'package.json'), 'utf8'));

/*
 * Locate `parse` by behavior, in a temporary module that re-exports the bundle's top-level
 * functions. Calling arbitrary bundle functions to probe them can throw asynchronously, so the
 * process-level handlers below keep an unrelated rejection from killing generation.
 */
process.on('uncaughtException', () => {});
process.on('unhandledRejection', () => {});

const distributionDirectory = NodePath.join(packageRoot, 'dist');
const bundleSource = NodeFileSystem.readFileSync(NodePath.join(distributionDirectory, 'lib.mjs'), 'utf8');

const declaredNames = [
    ...new Set(
        [...bundleSource.matchAll(/(?:^|[;}\s])function ([A-Za-z_$][A-Za-z0-9_$]*)\s*\(/g)].map((match) => match[1]),
    ),
];

/*
 * Not every match is a top-level binding — some are nested declarations that cannot be exported — so
 * each name is tried on its own and the ones that fail to link are dropped.
 */
const bindableNames = [];
for (const name of declaredNames) {
    const probePath = NodePath.join(distributionDirectory, `zz_gen_probe_${name}.mjs`);
    NodeFileSystem.writeFileSync(probePath, `${bundleSource}\nexport{${name} as probe};\n`);
    try {
        await import(NodeUrl.pathToFileURL(probePath).href);
        bindableNames.push(name);
    } catch {
        /* Not a top-level binding. */
    } finally {
        NodeFileSystem.unlinkSync(probePath);
    }
}

const patchedPath = NodePath.join(distributionDirectory, 'zz_gen_patched.mjs');
NodeFileSystem.writeFileSync(
    patchedPath,
    `${bundleSource}\nexport{${bindableNames.map((name) => `${name} as probe_${name}`).join(',')}};\n`,
);

let parse = null;
let parseExportName = null;

try {
    const patched = await import(NodeUrl.pathToFileURL(patchedPath).href);

    for (const exportName of Object.keys(patched)) {
        const exported = patched[exportName];
        if (typeof exported !== 'function') continue;

        try {
            const simple = exported('.a{color:red}');
            if (!Array.isArray(simple) || simple.length !== 1) continue;
            if (simple[0]?.kind !== 'rule' || simple[0]?.selector !== '.a') continue;

            // Separates the parser from the wrapper that yields a nested rule with a null body.
            const body = simple[0].nodes?.[0];
            if (body?.kind !== 'declaration' || body.property !== 'color' || body.value !== 'red') continue;

            // The custom-property branch keeps raw newlines; nothing else in the bundle does.
            const custom = exported(':root{--a: b,\n  c;}');
            const declaration = custom?.[0]?.nodes?.[0];
            if (declaration?.kind !== 'declaration') continue;
            if (typeof declaration.value !== 'string' || !declaration.value.includes('\n')) continue;

            parse = exported;
            parseExportName = exportName.replace(/^probe_/, '');
            break;
        } catch {
            /* Not the parser. */
        }
    }
} finally {
    NodeFileSystem.unlinkSync(patchedPath);
}

if (!parse) {
    process.stderr.write('could not locate parse() in the bundle by identity\n');
    process.exit(1);
}

/*
 * Serialize to exactly the fields the Go port models, matching generate_syntax_tree's shape so the two
 * fixtures read the same way. Source locations are dropped because the port does not carry them.
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
        default:
            throw new Error(`parse() produced an unexpected node kind: ${node.kind}`);
    }
}

/*
 * The corpus. Hand-written stylesheets covering each branch of the character loop, then every real
 * stylesheet named on the command line, so the suite carries both the constructed edge cases and
 * the actual repository input the exit criterion names.
 */
const syntheticCases = [
    ['empty', ''],
    ['whitespace only', '   \n\t\n  '],
    ['simple rule', '.a { color: red; }'],
    ['no trailing semicolon', '.a { color: red }'],
    ['multiple declarations', '.a { color: red; background: blue; }'],
    ['nested rule', '.a { .b { color: red; } }'],
    ['deeply nested', '.a { .b { .c { .d { color: red; } } } }'],
    ['at-rule with body', '@media (width >= 40rem) { .a { color: red; } }'],
    ['body-less at-rule', '@charset "UTF-8";'],
    ['import at-rule', '@import "tailwindcss";'],
    ['import with source', '@import "tailwindcss" source(none);'],
    ['at-rule no semicolon at eof', '@tailwind utilities'],
    ['at-rule closing a block', '@layer foo { @tailwind utilities }'],
    ['at-rule shorter than @page', '@x { color: red; }'],
    // The scan for an at-rule's name boundary starts at index 5, upstream's documented assumption
    // that `@page` is the shortest at-rule in CSS. These are the inputs where that is observable:
    // the engine reports `@x y` as the whole name with empty params, and a port that "fixed" the
    // scan to start at 1 would split them and disagree with the engine on every one.
    ['short at-rule with params', '@x y;'],
    ['short at-rule with params and body', '@x y { color: red; }'],
    ['short at-rule with paren params', '@md (a) { .b { c: d; } }'],
    ['page at-rule splits normally', '@page x;'],
    ['at-rule params with paren', '@media(width>=40rem){.a{color:red}}'],
    ['at-rule tab separated', '@media\t(width >= 40rem) { .a { color: red } }'],

    // Comments.
    ['comment dropped', '/* hello */ .a { color: red; }'],
    ['license comment hoisted', '/*! license */ .a { color: red; }'],
    ['license comment after rule', '.a { color: red; } /*! license */'],
    ['two license comments', '/*! one */ /*! two */ .a { color: red; }'],
    ['comment inside declaration', '.a { color: /* mid */ red; }'],
    ['comment with escaped star', '/* a \\* b */ .a { color: red; }'],
    ['unterminated comment', '.a { color: red; } /* never closes'],

    // Strings.
    ['single quoted', ".a { content: 'x'; }"],
    ['double quoted', '.a { content: "x"; }'],
    ['quote inside quote', `.a { content: "a 'b' c"; }`],
    ['semicolon inside string', '.a { content: "a;b"; }'],
    ['brace inside string', '.a { content: "a}b"; }'],
    ['colon inside string', '.a { content: "a:b"; }'],
    ['escaped quote in string', '.a { content: "a\\"b"; }'],

    // Escapes in selectors.
    ['escaped colon selector', '.hover\\:foo:hover { color: red; }'],
    ['escaped bracket selector', '.w-\\[10px\\] { width: 10px; }'],
    ['escaped slash selector', '.w-1\\/2 { width: 50%; }'],

    // Whitespace handling.
    ['newlines collapse', '.a {\n\n\n  color:\n\n red;\n}'],
    ['tabs collapse', '.a {\t\t\tcolor:\t\tred;\t}'],
    ['crlf line endings', '.a {\r\n  color: red;\r\n}'],
    ['crlf inside value', '.a {\r\n  grid-template:\r\n    "a"\r\n    "b";\r\n}'],
    ['multiline selector', '.a,\n.b,\n.c { color: red; }'],

    // Parens guarding semicolons and braces.
    ['semicolon in parens', '.a { background: url(a;b.png); }'],
    ['brace in parens', '.a { background: url(a{b}.png); }'],
    ['nested parens', '.a { width: calc((1px + 2px) * 3); }'],
    ['supports at-rule', '@supports (display: grid) { .a { color: red; } }'],
    ['paren in at-rule params', '@custom-variant dark (&:where(.dark, .dark *));'],
    ['custom variant with has', '@custom-variant dark (&:where(.d, .d *):not(:where(.l:not(:has(.d)), .l:not(:has(.d)) *)));'],

    // !important.
    ['important', '.a { color: red !important; }'],
    ['important no space', '.a { color: red!important; }'],
    ['important in custom property', '.a { --x: red !important; }'],

    // The custom-property branch. Every one of these is a case where treating `--foo` as an
    // ordinary declaration returns something plausible and wrong.
    ['custom property simple', ':root { --a: b; }'],
    ['custom property empty value', ':root { --a:; }'],
    ['custom property no semicolon', ':root { --a: b }'],
    ['custom property multiline', ':root {\n  --a: b,\n    c,\n    d;\n}'],
    ['custom property keeps indentation', ':root {\n  --font-sans:\n        var(--f), -apple-system, "Segoe UI",\n        sans-serif;\n}'],
    ['custom property with semicolon in parens', ':root { --a: url(x;y); }'],
    ['custom property with braces', ':root { --a: { b: c }; }'],
    ['custom property with brackets', ':root { --a: [b;c]; }'],
    ['custom property with nested brackets', ':root { --a: ([{;}]); }'],
    ['custom property with string', `:root { --a: "b;c"; }`],
    ['custom property with string brace', `:root { --a: "b}c"; }`],
    ['custom property with comment', ':root { --a: /* ; */ b; }'],
    ['custom property with colon in value', ':root { --a: url(http://x/y); }'],
    ['custom property with var fallback colon', ':root { --a: var(--b, url(http://x)); }'],
    ['custom property escaped', ':root { --a: \;b; }'],
    ['custom property at top level', '--a: b;'],
    ['custom property ends at brace', '.a { --x: y }'],
    ['custom property ends at eof', ':root { --a: b'],
    ['two custom properties', ':root { --a: b; --c: d; }'],
    ['custom property then declaration', ':root { --a: b; color: red; }'],
    ['custom property unbalanced open paren', ':root { --a: (b; }'],
    ['custom property tw namespace', '@theme { --color-a: #fff; --spacing: 0.25rem; }'],

    // @theme and @utility shapes, the ones this port exists to read.
    ['theme block', '@theme { --color-red-500: oklch(0.637 0.237 25.331); }'],
    ['theme inline', '@theme inline { --animate-x: x 1s; }'],
    ['theme with keyframes', '@theme { --animate-spin: spin 1s linear infinite; @keyframes spin { to { transform: rotate(360deg); } } }'],
    ['utility static', '@utility scrollbar-hide { scrollbar-width: none; }'],
    ['utility functional', '@utility border--* { border-color: --value(--color- *); }'],
    ['utility negative', '@utility -zoom-in-* { --tw-enter-scale: calc(--value([integer]) * 1%); }'],
    ['utility with nested selector', '@utility prose { p { margin: 1em 0; } }'],
    ['variant at-rule', '@variant dark { color: white; }'],
    ['apply directive', '.a { @apply flex items-center; }'],
    ['layer block', '@layer base { strong { font-weight: 500; } }'],
    ['config at-rule', "@config '../TailwindConfiguration.ts';"],
    ['source at-rule', "@source '../app/';"],

    // Byte-order mark.
    ['byte order mark', '﻿.a { color: red; }'],
];

const cases = [];
const errorCases = [];

for (const [name, input] of syntheticCases) {
    try {
        cases.push({ name, source: 'synthetic', input, nodes: parse(input).map(serializeNode) });
    } catch (error) {
        errorCases.push({ name, source: 'synthetic', input, message: error.message });
    }
}

/*
 * Stylesheets that must fail. A parser that accepts these is not stricter, it is wrong: the engine
 * rejects them and any tree the port invented would not be a tree the engine ever produces.
 */
const errorInputs = [
    ['missing opening brace', '.a { color: red; } }'],
    ['missing opening paren', '.a { width: 1px); }'],
    ['missing closing brace rule', '.a { color: red;'],
    ['missing closing brace at-rule', '@media (width >= 40rem) { .a { color: red; }'],
    ['invalid declaration', '.a { nonsense; }'],
    ['unterminated string', '.a { content: "abc\n}'],
    ['unterminated string with semicolon', '.a { content: "abc;\n}'],
    ['custom property without value', '.a { --abc }'],
];

for (const [name, input] of errorInputs) {
    try {
        const nodes = parse(input).map(serializeNode);
        errorCases.push({ name, source: 'synthetic-error', input, message: null, unexpectedlyParsed: nodes });
    } catch (error) {
        errorCases.push({ name, source: 'synthetic-error', input, message: error.message });
    }
}

/*
 * Real stylesheets. The exit criterion is that the port reproduces what the engine sees for an
 * actual repository's theme and its whole @import graph, not for constructed examples.
 */
const realFiles = corpusPathArgument
    ? JSON.parse(NodeFileSystem.readFileSync(NodePath.resolve(corpusPathArgument), 'utf8'))
    : [];

let themeEntryCount = 0;
let utilityBlockCount = 0;
let customVariantCount = 0;

function countAtRules(nodes) {
    for (const node of nodes) {
        if (node.kind === 'at-rule') {
            if (node.name === '@theme') {
                for (const child of node.nodes) {
                    if (child.kind === 'declaration' && child.property.startsWith('--')) themeEntryCount++;
                }
            }
            if (node.name === '@utility') utilityBlockCount++;
            if (node.name === '@custom-variant') customVariantCount++;
        }
        if (node.nodes) countAtRules(node.nodes);
    }
}

for (const entry of realFiles) {
    const content = NodeFileSystem.readFileSync(entry.path, 'utf8');
    const nodes = parse(content).map(serializeNode);
    countAtRules(nodes);
    cases.push({ name: entry.name, source: 'repository', input: content, nodes });
}

process.stdout.write(
    JSON.stringify(
        {
            tailwindVersion: packageJson.version,
            parseExportName,
            parseResolvedByIdentity: true,
            stylesheetCount: realFiles.length,
            themeEntryCount,
            utilityBlockCount,
            customVariantCount,
            cases,
            errorCases,
        },
        null,
        2,
    ) + '\n',
);
