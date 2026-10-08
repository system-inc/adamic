// Type-only adaptations. Reviewed recovery branches justify the object views.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const {aliases, rules} = require('./rules.json');
if (ts.version !== '6.0.3') throw Error('requires stock TypeScript 6.0.3');
if (process.argv.length !== 3) throw Error('usage: node adapt.cjs <tree>');
const root = path.resolve(process.argv[2]);
const texts = new Map();
for (const rule of rules) {
    const file = path.join(root, rule.file);
    let text = texts.get(file) ?? fs.readFileSync(file, 'utf8');
    const count = text.split(rule.before).length - 1;
    if (count === 1) text = text.replace(rule.before, rule.after);
    else if (count || text.split(rule.after).length !== 2) throw Error('unreviewed site: ' + rule.name);
    texts.set(file, text);
}
const config = path.join(root, 'src/compiler/commandLineParser.ts');
if (!texts.get(config).includes(aliases)) {
    texts.set(config, texts.get(config).replace('function convertConfigFileToObject(', aliases + 'function convertConfigFileToObject('));
}
// Validate all edits before writing any file.
for (const [file, text] of texts) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw Error('invalid adaptation: ' + file);
}
for (const [file, text] of texts) if (text !== fs.readFileSync(file, 'utf8')) fs.writeFileSync(file, text);
console.log(JSON.stringify({adapted: 6, declined: ['convertToObject', 'tryParseJson']}));
