// Independent oracle: the exact upstream release cohere follows.
import { createRequire } from 'node:module';
import { readFileSync, readdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
const [directory, mode, root, destination, testTexts] = process.argv.slice(2);
const require = createRequire(join(directory, 'package.json'));
if(require('postcss-selector-parser/package.json').version !== '2.2.3') {
    throw new Error('expected postcss-selector-parser 2.2.3');
}
const Processor = require('postcss-selector-parser/dist/processor.js');
const postcss = require('postcss');
const decode = (line) => line.slice(1).replace(/\\(.)/g, (_, c) => ({ '\\': '\\', t: '\t', n: '\n', r: '\r' })[c]);
function byteOffset(text, index) {
    let units = 0,
        bytes = 0;
    for(const c of text) {
        if(index < units + c.length) return bytes;
        units += c.length;
        bytes += Buffer.byteLength(c);
    }
    return bytes + index - units;
}
function dump(value, text) {
    if(value === undefined) return '<undefined>';
    if(typeof value === 'string') return value.toWellFormed();
    if(Array.isArray(value)) return value.map((v) => dump(v, text));
    if(value === null || typeof value !== 'object') return value;
    const out = {};
    for(const key of Object.keys(value).sort()) {
        if(key === 'parent') continue;
        out[key] =
            key === 'sourceIndex' && typeof value[key] === 'number' && Number.isFinite(value[key])
                ? byteOffset(text, value[key])
                : dump(value[key], text);
    }
    return out;
}
const loops = 'postcss-selector-parser 2.2.3 never returns on this selector: a namespace bar it does not consume';
if(mode === 'corpus') {
    let files = 0;
    const selectors = [];
    function extract(text, file) {
        postcss.parse(text, { from: file }).walk((node) => {
            if(typeof node.selector === 'string') {
                let selector = node.raws.selector?.scss ?? node.raws.selector?.raw ?? node.selector;
                if(node.raws.between?.trim()) selector += node.raws.between;
                if(selector.trim()) selectors.push(selector);
            }
            if(node.type === 'atrule') {
                let params = node.raws.params?.scss ?? node.raws.params?.raw ?? node.params;
                if(node.raws.afterName?.trim()) params = node.raws.afterName + params;
                if(node.raws.between?.trim()) params += node.raws.between;
                if(
                    ['extend', 'nest', 'at-root'].includes(node.name) &&
                    params.trim() &&
                    !(node.name === 'at-root' && /^\(\s*(?:without|with)\s*:.+\)$/s.test(params))
                )
                    selectors.push(params.trim());
                if(node.name === 'custom-selector') {
                    const match = node.params.match(/:--\S+\s+/);
                    if(match) selectors.push(node.params.slice(match[0].trim().length).trim());
                }
            }
        });
    }
    function walk(path) {
        for(const entry of readdirSync(path, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
            if(entry.name === '.git' || entry.name === 'node_modules') continue;
            const file = join(path, entry.name);
            if(entry.isDirectory()) {
                walk(file);
                continue;
            }
            if(!entry.name.endsWith('.css')) continue;
            files++;
            // A malformed CSS file is a corpus failure, never quietly omitted.
            extract(readFileSync(file, 'utf8'), file);
        }
    }
    walk(root);
    const fileSelectors = selectors.length;
    let candidates = 0,
        invalidSnippets = 0;
    if(testTexts) {
        for(const text of JSON.parse(readFileSync(testTexts, 'utf8'))) {
            if(!text.includes('{')) continue;
            candidates++;
            try {
                extract(text, '<Go CSS test constant>');
            }
            catch {
                invalidSnippets++;
            }
        }
    }
    console.log(
        `${candidates} brace-containing CSS test constants, ${invalidSnippets} not valid stylesheets, ${selectors.length - fileSelectors} extracted test selectors`,
    );
    writeFileSync(destination, selectors.map((s) => JSON.stringify(s)).join('\n') + '\n');
    console.log(`${files} CSS files, ${selectors.length} selectors (including repeats)`);
}
else if(mode === 'count') {
    const texts = readFileSync(root, 'utf8').split('\n').filter(Boolean).map(decode);
    const answers = readFileSync(destination, 'utf8').split('\n');
    let parsed = 0,
        nodes = 0;
    for(let round = 0; round < 10; round++) {
        texts.forEach((text, index) => {
            if(answers[index * 2 + 1] === `error ${loops}`) return;
            try {
                let tree;
                new Processor((value) => {
                    tree = value;
                }).process(text);
                parsed++;
                nodes += tree.nodes.length;
            }
            catch {
                /* A refusal is part of the measured workload. */
            }
        });
    }
    console.log(`${parsed} of ${texts.length * 10} selectors parsed, ${nodes} nodes`);
}
else {
    const texts = readFileSync(mode, 'utf8').split('\n').filter(Boolean).map(decode);
    // Go's answers identify only its deliberate nontermination refusal. These
    // are excluded from JS execution, counted explicitly, never called matches.
    const answers = root ? readFileSync(root, 'utf8').split('\n') : [];
    const out = [];
    let excluded = 0;
    texts.forEach((text, index) => {
        out.push(`case ${index}`);
        if(answers[index * 2 + 1] === `error ${loops}`) {
            out.push(`error ${loops}`);
            excluded++;
            return;
        }
        try {
            let tree;
            new Processor((value) => {
                tree = value;
            }).process(text);
            out.push(JSON.stringify(dump(tree, text)).replaceAll('\u2028', '\\u2028').replaceAll('\u2029', '\\u2029'));
        }
        catch(error) {
            out.push(`error ${error.message}`);
        }
    });
    process.stdout.write(out.join('\n') + '\n');
    if(excluded) process.stderr.write(`${excluded} nonterminating inputs excluded from JS oracle\n`);
}
