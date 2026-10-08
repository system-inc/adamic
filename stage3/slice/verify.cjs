#!/usr/bin/env node
// Audit copied UTF-16 spans as UTF-8 bytes, including namespace wrapper pieces.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const root = path.resolve(process.argv[2]);
const manifest = JSON.parse(fs.readFileSync(path.join(root,'slice.json'),'utf8'));
const sources = new Map(), outputs = new Map();
for (const record of manifest.declarations) {
    if (!sources.has(record.file)) sources.set(record.file,fs.readFileSync(path.join(manifest.summary.tree,record.file),'utf8'));
    if (!outputs.has(record.file)) outputs.set(record.file,fs.readFileSync(path.join(root,manifest.adaptations ? '.verbatim' : '',record.file),'utf8'));
    const original = Buffer.from(sources.get(record.file).slice(record.start,record.end));
    const copied = Buffer.from(outputs.get(record.file).slice(record.output_start,record.output_end));
    if (!original.equals(copied) || copied.length !== record.bytes
        || crypto.createHash('sha256').update(copied).digest('hex') !== record.sha256) {
        throw new Error(`copied bytes differ: ${record.file}:${record.line} ${record.names.join(',')}`);
    }
}
for (const module of manifest.evaluation || []) {
    const destination = path.join(root, module.file);
    const text = outputs.get(module.file) || fs.readFileSync(destination, 'utf8');
    const lines = text.split(/\r?\n/);
    for (let index = 0; index < module.imports.length; index++) {
        let specifier = path.relative(path.dirname(destination), path.join(root, module.imports[index])).replaceAll(path.sep, '/');
        if (!specifier.startsWith('.')) specifier = './' + specifier;
        if (lines[index] !== `import ${JSON.stringify(specifier)};`) throw new Error(`evaluation order differs: ${module.file}:${index + 1}`);
    }
}
console.log(`PASS: ${manifest.declarations.length} byte-identical source spans; ${(manifest.evaluation || []).length} ordered module import lists`);

if (manifest.adaptations) for (const [file, hash] of Object.entries(manifest.adaptations.final_sha256)) {
    if (crypto.createHash('sha256').update(fs.readFileSync(path.join(root,file))).digest('hex') !== hash) throw new Error('adapted bytes differ: ' + file);
}
