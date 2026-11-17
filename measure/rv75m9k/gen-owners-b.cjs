#!/usr/bin/env node
'use strict';
// #rv75m9k road B: owners.json for the receiver printer's engine (engine-adapt.cjs), from a tree before and after the
// codemod. An owner is a named function (declaration or method), identified as the engine identifies it: its kind, its
// name and the names of the function-likes and modules around it, itself included. Every changed owner is found; an
// outer one is pinned only when its own text changed outside the inner ones (an arrow made once at its top), and an
// owner inside a pinned one is marked coveredBy it, so the engine checks its pin and writes nothing for it. A change
// outside every named function, or an owner the engine would find ambiguously, is refused before anything is written.
// Usage: NODE_PATH=... node gen-owners-b.cjs <before src/compiler dir> <after tree> <owners.json>
const ts = require('typescript'), fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
const [beforeDirectory, tree, out] = process.argv.slice(2);
const owners = [], refusals = [];

function ownerPathOf(node, source) {
    const names = [];
    for (let current = node; current; current = current.parent)
        if ((ts.isFunctionLike(current) || ts.isModuleDeclaration(current)) && current.name) names.unshift(current.name.getText(source));
    return names;
}
function namedOwners(source) {
    const found = [];
    (function visit(node) {
        // Overload signatures too: the engine counts every declaration of the name, and pins them as one span.
        if ((ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node)) && node.name) found.push(node);
        ts.forEachChild(node, visit);
    })(source);
    return found;
}
function identity(node, source) { return `${ts.SyntaxKind[node.kind]}|${node.name.getText(source)}|${JSON.stringify(ownerPathOf(node, source))}`; }

const changedFiles = [];
(function walk(directory, relative) {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
        const full = path.join(directory, entry.name), name = path.join(relative, entry.name);
        if (entry.isDirectory()) walk(full, name);
        else if (name.endsWith('.ts') && fs.readFileSync(full, 'utf8') !== fs.readFileSync(path.join(tree, 'src/compiler', name), 'utf8')) changedFiles.push(name);
    }
})(beforeDirectory, '');

for (const name of changedFiles.sort()) {
    const file = 'src/compiler/' + name;
    const rawBefore = fs.readFileSync(path.join(beforeDirectory, name), 'utf8'), rawAfter = fs.readFileSync(path.join(tree, file), 'utf8');
    // The engine pins LF text and restores the file's own ending.
    const before = rawBefore.replace(/\r\n/g, '\n'), after = rawAfter.replace(/\r\n/g, '\n');
    const sourceBefore = ts.createSourceFile(file, before, ts.ScriptTarget.Latest, true), sourceAfter = ts.createSourceFile(file, after, ts.ScriptTarget.Latest, true);
    const groups = (source, text) => {
        const map = new Map();
        for (const node of namedOwners(source)) { const key = identity(node, source); if (!map.has(key)) map.set(key, []); map.get(key).push(node); }
        const spans = new Map();
        for (const [key, nodes] of map) {
            nodes.sort((a, b) => a.pos - b.pos);
            spans.set(key, { nodes, start: nodes[0].getStart(source), end: nodes.at(-1).end, text: text.slice(nodes[0].getStart(source), nodes.at(-1).end) });
        }
        return spans;
    };
    const beforeGroups = groups(sourceBefore, before), afterGroups = groups(sourceAfter, after);
    const changed = [...afterGroups.keys()].filter(key => beforeGroups.has(key) && beforeGroups.get(key).text !== afterGroups.get(key).text);
    for (const key of afterGroups.keys()) if (!beforeGroups.has(key)) refusals.push(`${file}: an owner only the after tree holds, ${key}`);
    // Innermost first: an outer owner is needed only if its text, with every selected inner owner put back to its
    // before text, still differs from its before text.
    const selected = [];
    const depth = key => JSON.parse(key.split('|')[2]).length;
    for (const key of changed.sort((a, b) => depth(b) - depth(a))) {
        const span = afterGroups.get(key);
        let text = span.text;
        const inner = selected.filter(other => other.file === file && afterGroups.get(other.key).start >= span.start && afterGroups.get(other.key).end <= span.end)
            .sort((a, b) => afterGroups.get(b.key).start - afterGroups.get(a.key).start);
        for (const other of inner) {
            const otherSpan = afterGroups.get(other.key);
            text = text.slice(0, otherSpan.start - span.start) + beforeGroups.get(other.key).text + text.slice(otherSpan.end - span.start);
        }
        if (text === beforeGroups.get(key).text) continue;
        selected.push({ file, key });
    }
    // A change outside every selected owner can't be pinned by this engine.
    let rebuilt = after;
    for (const owner of [...selected].sort((a, b) => afterGroups.get(b.key).start - afterGroups.get(a.key).start)) {
        const span = afterGroups.get(owner.key);
        if (selected.some(other => other !== owner && afterGroups.get(other.key).start <= span.start && afterGroups.get(other.key).end >= span.end)) continue;
        rebuilt = rebuilt.slice(0, span.start) + beforeGroups.get(owner.key).text + rebuilt.slice(span.end);
    }
    if (rebuilt !== before) refusals.push(`${file}: a change outside every named function`);
    for (const owner of selected) {
        const [kind, ownerName, ownerPath] = owner.key.split('|');
        const span = afterGroups.get(owner.key), beforeText = beforeGroups.get(owner.key).text;
        const outer = selected.find(other => other !== owner && afterGroups.get(other.key).start <= span.start && afterGroups.get(other.key).end >= span.end);
        owners.push({
            file, owner: ownerName, kind, ownerPath: JSON.parse(ownerPath), declarations: span.nodes.length,
            ...(outer ? { coveredBy: outer.key.split('|')[1] } : {}),
            before: beforeText, after: span.text, ownerBeforeSha256: digest(beforeText), ownerAfterSha256: digest(span.text),
        });
    }
}
if (refusals.length) { for (const refusal of refusals) console.log('refused ' + refusal); process.exit(1); }
fs.writeFileSync(out, JSON.stringify(owners, null, 2) + '\n');
console.log(`owners: ${owners.length} owners in ${new Set(owners.map(owner => owner.file)).size} files, ${owners.filter(owner => owner.coveredBy).length} covered`);
for (const owner of owners) console.log(`owner ${owner.file} ${owner.kind} ${owner.ownerPath.join('.')}${owner.coveredBy ? ' (covered by ' + owner.coveredBy + ')' : ''} ${owner.before.length} -> ${owner.after.length}`);
