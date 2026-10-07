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
    if (!outputs.has(record.file)) outputs.set(record.file,fs.readFileSync(path.join(root,record.file),'utf8'));
    const original = Buffer.from(sources.get(record.file).slice(record.start,record.end));
    const copied = Buffer.from(outputs.get(record.file).slice(record.output_start,record.output_end));
    if (!original.equals(copied) || copied.length !== record.bytes
        || crypto.createHash('sha256').update(copied).digest('hex') !== record.sha256) {
        throw new Error(`copied bytes differ: ${record.file}:${record.line} ${record.names.join(',')}`);
    }
}
console.log(`PASS: ${manifest.declarations.length} byte-identical source spans`);
