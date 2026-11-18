'use strict';
// #0b9xqhh's census: node rewrite72.cjs TREE SCAN.json OUT.json [EXCLUDE.json]
// Spells every family 5 relation the scan stopped on (refuseOptionalWidening, "absent from structural source")
// as a cast to its slot's own type, `(expr as (Slot))`, in place in TREE, so the census prototype's checked view
// (rulings.md:72 at a cast) judges it. Writes OUT.json: each site, the expression, the slot type spelled, and
// why a site was left alone. EXCLUDE.json lists sites (file:line:col) a first pass found the spelling broke.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const [tree, scanFile, outFile, excludeFile] = process.argv.slice(2);
const exclude = new Set(excludeFile && fs.existsSync(excludeFile) ? JSON.parse(fs.readFileSync(excludeFile, 'utf8')) : []);
const scan = JSON.parse(fs.readFileSync(scanFile, 'utf8'));
const family5 = scan.Stops.filter(stop => stop.Words.includes('absent from structural source'));

// Every family 5 stop, by its rule: only refuseOptionalWidening's sites are values in a slot; the others
// (a type argument against a constraint, constraintRelationRefusal) have no expression to cast, named as such.
const sites = [];
for (const stop of family5) {
    const match = /^(.*):(\d+):(\d+)$/.exec(stop.Site);
    const rel = match ? match[1].split('/tree-').pop().replace(/^[^/]*\//, '') : stop.Site;
    sites.push({site: stop.Site, file: match ? match[1] : '', line: match ? +match[2] : 0, column: match ? +match[3] : 0, rule: stop.Rule.split('/').pop(), words: stop.Words, relative: rel});
}

// Adamic's one configuration (internal/load/load.go compilerOptions), over the scan's own root, so a tree this
// calls clean is one the scan's checker loads.
const adamicOptions = {strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true, erasableSyntaxOnly: false, verbatimModuleSyntax: true, allowImportingTsExtensions: true, noEmit: true, module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force, moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024, lib: ['lib.es2024.d.ts'], types: []};
const program = ts.createProgram([path.join(tree, 'src/tsc/tsc.ts')], adamicOptions);
const checker = program.getTypeChecker();

// The value in a slot whose start the stop names: the outermost expression starting there.
function locate(source, line, column) {
    let position;
    try { position = source.getPositionOfLineAndCharacter(line - 1, column - 1); } catch { return undefined; }
    let found;
    (function visit(node) {
        if (node.getStart(source) === position && ts.isExpression(node) && !found) { found = node; return; }
        if (node.pos <= position && position < node.end) ts.forEachChild(node, visit);
    })(source);
    while (found && found.parent && ts.isExpression(found.parent) && found.parent.getStart(source) === position && !ts.isParenthesizedExpression(found.parent)) found = found.parent;
    return found;
}

const edits = new Map(); // file -> [{start, end, text}]
for (const site of sites) {
    if (site.rule !== 'refuseOptionalWidening') { site.left = `rule ${site.rule}: no value to cast`; continue; }
    const key = `${site.relative}:${site.line}:${site.column}`;
    site.key = key;
    if (exclude.has(key)) { site.left = 'the cast spelling broke stock tsc on the first pass'; continue; }
    const file = path.join(tree, site.relative);
    const source = program.getSourceFile(file);
    if (!source) { site.left = `no source file ${site.relative}`; continue; }
    const node = locate(source, site.line, site.column);
    if (!node) { site.left = 'no expression starts at the site'; continue; }
    const slot = checker.getContextualType(node);
    if (!slot) { site.left = 'no contextual slot type'; continue; }
    const spelled = checker.typeToString(slot, node, ts.TypeFormatFlags.NoTruncation | ts.TypeFormatFlags.UseFullyQualifiedType | ts.TypeFormatFlags.WriteArrowStyleSignature);
    if (/\.\.\.|any\b/.test(spelled) && /\.\.\./.test(spelled)) { site.left = `slot type prints truncated: ${spelled.slice(0, 80)}`; continue; }
    site.expression = node.getText(source).slice(0, 160);
    site.slot = spelled;
    const list = edits.get(file) || [];
    if (list.some(edit => edit.start === node.getStart(source))) { site.left = 'a second stop at the same expression'; continue; }
    list.push({start: node.getStart(source), end: node.end, slot: spelled, key});
    edits.set(file, list);
    site.rewritten = true;
    site.start = node.getStart(source);
    site.end = node.end;
}
for (const [file, list] of edits) {
    let text = fs.readFileSync(file, 'utf8');
    // A site nested inside another site's expression is left to the outer cast's view, named, so no two
    // edits overlap and each applies from the end of the file back.
    for (const inner of list) for (const outer of list) if (outer !== inner && outer.start <= inner.start && inner.end <= outer.end && !(outer.start === inner.start && outer.end === inner.end)) inner.nested = true;
    for (const site of sites) if (site.rewritten && list.some(edit => edit.nested && edit.key === site.key)) { site.rewritten = false; site.left = 'nested inside another site\'s expression'; }
    list.splice(0, list.length, ...list.filter(edit => !edit.nested));
    list.sort((a, b) => b.start - a.start);
    for (const edit of list) text = text.slice(0, edit.start) + '(' + text.slice(edit.start, edit.end) + ' as (' + edit.slot + '))' + text.slice(edit.end);
    fs.writeFileSync(file, text);
}
fs.writeFileSync(outFile, JSON.stringify(sites, null, 1));
const count = (f) => sites.filter(f).length;
console.log(`rewrite72: ${sites.length} family 5 stops, ${count(s => s.rewritten)} rewritten, ${count(s => s.left)} left alone`);
