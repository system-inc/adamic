/*
 * Ask the real Tailwind value parser what tree every shape in the corpus produces, and emit the
 * answers as JSON.
 *
 * The Go port in internal/tailwind is checked against these trees rather than against a reading of
 * `value-parser.ts`, because several of this parser's behaviors are consequences of JavaScript
 * semantics rather than of stated intent, and a careful reader transcribing the source would get
 * them wrong in the direction of a plausible tree. Three that this corpus pins:
 *
 *   - `a)b` parses to just `[word("b")]`. The close-paren branch flushes the pending buffer with
 *     `tail?.nodes.push(node)`, and when the stack is empty `tail` is undefined, so the optional
 *     chaining silently discards the word. `foo\(bar)` likewise parses to `[]`.
 *   - `\` at the end of input produces the word `\undefined`, literally. The backslash branch is
 *     `buffer += input[i] + input[i + 1]`, and indexing one past the end yields `undefined`, which
 *     string-concatenates rather than throwing.
 *   - `foo(bar` leaves the function node with no children and puts `bar` at the top level. The
 *     final buffer flush pushes to `ast`, never to the still-open `parent`.
 *
 * # Reaching the parser
 *
 * Unlike `inferDataType`, this function is not exported from any dist chunk. The published package
 * ships no `src/`, and the bundler inlined the value parser into `lib.mjs` without re-exporting it,
 * so there is nothing to `import`. It is instead located by structure and extracted as text.
 *
 * The anchor is the three node constructors, which the minifier preserved as recognizable object
 * literals: `{kind:"word",value:X}`, `{kind:"function",value:X,nodes:Y}`, `{kind:"separator",
 * value:X}`. Those literal keys survive minification because they are data, not identifiers. The
 * parse function is the next `function F(e){e=e.replaceAll(...)` after them. The extracted text is
 * then evaluated and identity-checked against behaviors no other function in the bundle has, so a
 * bundler change that moves things produces a loud failure rather than a fixture generated from the
 * wrong function.
 *
 * Nothing here is hardcoded to a chunk hash or a mangled name; both change on every release.
 */

import * as NodeFileSystem from 'node:fs';
import * as NodePath from 'node:path';
import * as NodeUrl from 'node:url';
import * as NodeOs from 'node:os';

const [, , packageRootArgument] = process.argv;
if (!packageRootArgument) {
    process.stderr.write('usage: enumerate.mjs <tailwindcss package root>\n');
    process.exit(2);
}

const packageRoot = NodePath.resolve(packageRootArgument);
const packageJson = JSON.parse(NodeFileSystem.readFileSync(NodePath.join(packageRoot, 'package.json'), 'utf8'));

const bundlePath = NodePath.join(packageRoot, 'dist', 'lib.mjs');
const bundleSource = NodeFileSystem.readFileSync(bundlePath, 'utf8');

/*
 * Locate the three node constructors. They sit adjacent in the bundle because they are adjacent in
 * the source, and each is small enough that the minifier left the shape intact.
 */
const wordConstructor = /function ([A-Za-z_$][\w$]*)\(([A-Za-z_$][\w$]*)\)\{return\{kind:"word",value:\2\}\}/.exec(bundleSource);
const functionConstructor = /function ([A-Za-z_$][\w$]*)\(([A-Za-z_$][\w$]*),([A-Za-z_$][\w$]*)\)\{return\{kind:"function",value:\2,nodes:\3\}\}/.exec(bundleSource);
const separatorConstructor = /function ([A-Za-z_$][\w$]*)\(([A-Za-z_$][\w$]*)\)\{return\{kind:"separator",value:\2\}\}/.exec(bundleSource);

if (!wordConstructor || !functionConstructor || !separatorConstructor) {
    process.stderr.write('could not locate the value-parser node constructors in dist/lib.mjs\n');
    process.exit(1);
}

/*
 * The parse function is the first one after the constructors whose body opens by normalising CRLF.
 * That first statement is unique to `parse` in this bundle and is the reason the search starts from
 * the constructors rather than from the top of the file.
 */
const afterConstructors = bundleSource.slice(wordConstructor.index);
const parseDeclaration = /function ([A-Za-z_$][\w$]*)\(([A-Za-z_$][\w$]*)\)\{\2=\2\.replaceAll\(/.exec(afterConstructors);
if (!parseDeclaration) {
    process.stderr.write('could not locate the parse function after the node constructors\n');
    process.exit(1);
}

/*
 * Slice from the first constructor through the end of the parse function by balancing braces, with
 * string literals skipped so that a brace inside a template or a quoted character does not throw
 * the count off. Regex literals are not a hazard here: this span contains none.
 */
function endOfFunctionBody(source, searchFrom) {
    const bodyStart = source.indexOf('{', searchFrom);
    let depth = 0;
    let openQuote = null;
    for (let index = bodyStart; index < source.length; index++) {
        const character = source[index];
        if (openQuote !== null) {
            if (character === '\\') {
                index++;
                continue;
            }
            if (character === openQuote) openQuote = null;
            continue;
        }
        if (character === '"' || character === "'" || character === '`') {
            openQuote = character;
            continue;
        }
        if (character === '{') depth++;
        else if (character === '}') {
            depth--;
            if (depth === 0) return index;
        }
    }
    return -1;
}

const parseAbsoluteIndex = wordConstructor.index + parseDeclaration.index;
/*
 * The brace search starts at the head of the declaration, not at the end of the matched text. The
 * matched text runs past the function's own opening brace and into the body, so searching from its
 * end would find the `for` loop's brace and balance one level too shallow, truncating the slice
 * just before the closing `return`.
 */
const bodyEnd = endOfFunctionBody(bundleSource, parseAbsoluteIndex);
if (bodyEnd === -1) {
    process.stderr.write('could not find the end of the parse function body\n');
    process.exit(1);
}

/*
 * `toCss` is also wanted, so that round-tripping can be measured rather than assumed. It is the
 * function taking an ast and concatenating, and it sits between the constructors and parse.
 */
const toCssDeclaration = /function ([A-Za-z_$][\w$]*)\(([A-Za-z_$][\w$]*)\)\{let ([A-Za-z_$][\w$]*)="";for\(let ([A-Za-z_$][\w$]*) of \2\)switch\(\4\.kind\)/.exec(afterConstructors);
if (!toCssDeclaration) {
    process.stderr.write('could not locate toCss after the node constructors\n');
    process.exit(1);
}

const moduleSource = bundleSource.slice(wordConstructor.index, bodyEnd + 1)
    + `\nexport const parse = ${parseDeclaration[1]};`
    + `\nexport const toCss = ${toCssDeclaration[1]};\n`;

const temporaryModulePath = NodePath.join(
    NodeFileSystem.mkdtempSync(NodePath.join(NodeOs.tmpdir(), 'tailwind-value-parser-')),
    'extracted.mjs',
);
NodeFileSystem.writeFileSync(temporaryModulePath, moduleSource);

const { parse, toCss } = await import(NodeUrl.pathToFileURL(temporaryModulePath).href);

/*
 * Identity check. Every assertion below is a behavior of the value parser specifically, and the
 * last three are the JavaScript-semantics consequences that no reimplementation would exhibit by
 * accident. If the bundler ever hands this extraction a different function, this fails loudly
 * instead of producing a fixture that quietly tests the wrong thing.
 */
function assertIdentity(description, actual, expected) {
    const actualJson = JSON.stringify(actual);
    const expectedJson = JSON.stringify(expected);
    if (actualJson !== expectedJson) {
        process.stderr.write(`extracted function failed identity check (${description})\n  got:  ${actualJson}\n  want: ${expectedJson}\n`);
        process.exit(1);
    }
}

assertIdentity('nested function call', parse('foo(bar, baz)'), [
    { kind: 'function', value: 'foo', nodes: [
        { kind: 'word', value: 'bar' },
        { kind: 'separator', value: ', ' },
        { kind: 'word', value: 'baz' },
    ] },
]);
assertIdentity('slash is its own word', parse('a/b'), [
    { kind: 'word', value: 'a' },
    { kind: 'word', value: '/' },
    { kind: 'word', value: 'b' },
]);
assertIdentity('buffer before an unmatched close paren is dropped', parse('a)b'), [
    { kind: 'word', value: 'b' },
]);
assertIdentity('trailing backslash reads one past the end', parse('\\'), [
    { kind: 'word', value: '\\undefined' },
]);
assertIdentity('remainder of an unclosed function lands at top level', parse('foo(bar'), [
    { kind: 'function', value: 'foo', nodes: [] },
    { kind: 'word', value: 'bar' },
]);
if (toCss(parse('calc(var(--a) * 2)')) !== 'calc(var(--a) * 2)') {
    process.stderr.write('extracted toCss failed its identity check\n');
    process.exit(1);
}

const corpus = JSON.parse(NodeFileSystem.readFileSync(new URL('./corpus.json', import.meta.url), 'utf8'));

/*
 * Beyond the hand-written corpus, enumerate every string of length 1 to 3 over an alphabet chosen
 * to stress the branches that interact: the escape, both quote kinds, both parens, separators, the
 * slash, and multi-byte characters of two, three and four UTF-8 bytes.
 *
 * The hand-written corpus names shapes someone thought of. This names shapes nobody thought of, and
 * it is what settles questions the corpus cannot. Whether the escape branch consumes one byte or
 * one whole character is invisible on every value in the corpus, and exhaustion over this alphabet
 * is what shows the two rules agree on all of them rather than merely on the ones tried.
 */
const exhaustiveAlphabet = ['\\', 'a', ' ', ',', '(', ')', '"', "'", '/', ':', 'é', '日', '\u{1F600}'];
const exhaustive = [];
(function enumerateStrings(prefix, remaining) {
    if (remaining === 0) {
        if (prefix !== '') exhaustive.push(prefix);
        return;
    }
    for (const character of exhaustiveAlphabet) enumerateStrings(prefix + character, remaining - 1);
})('', 1);
for (let length = 2; length <= 3; length++) {
    (function enumerateStrings(prefix, remaining) {
        if (remaining === 0) {
            exhaustive.push(prefix);
            return;
        }
        for (const character of exhaustiveAlphabet) enumerateStrings(prefix + character, remaining - 1);
    })('', length);
}

const cases = [];
for (const value of corpus) {
    const ast = parse(value);
    cases.push({
        value,
        ast,
        /*
         * The printed form is captured alongside the tree so the port is held to round-tripping as
         * well as to structure. They are different claims: a tree that drops a separator and a tree
         * that merges two words can both print correctly, and a tree that is structurally right can
         * still print wrong if whitespace moved between nodes.
         */
        css: toCss(ast),
    });
}

const exhaustiveCases = [];
for (const value of exhaustive) {
    const ast = parse(value);
    exhaustiveCases.push({ value, ast, css: toCss(ast) });
}

process.stdout.write(JSON.stringify({
    tailwindVersion: packageJson.version,
    bundle: NodePath.basename(bundlePath),
    parseSymbol: parseDeclaration[1],
    toCssSymbol: toCssDeclaration[1],
    exhaustiveAlphabet,
    cases,
    exhaustiveCases,
}, null, 2) + '\n');
