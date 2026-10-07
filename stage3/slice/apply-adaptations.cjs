#!/usr/bin/env node
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const {execFileSync} = require('node:child_process');
function apply(root) {
    root = path.resolve(root);
    const manifestPath = path.join(root, 'slice.json');
    const manifest = JSON.parse(fs.readFileSync(manifestPath));
    if (manifest.adaptations) throw new Error('slice adaptations already applied');
    // Parser temporaries 60–64 are full-tree edits, already present in the input.
    const scanner = manifest.summary.entries.some(e => e.endsWith(':createScanner')) && !manifest.summary.entries.some(e => e.endsWith(':createSourceFile'));
    const numbers = scanner ? [52,53,54,55,56,57,59,80,81,82,85,88,89] : [];
    if (!numbers.length) return;
    execFileSync(process.execPath, [path.join(__dirname,'verify.cjs'),root], {stdio:'inherit'});
    const files = [...new Set(manifest.declarations.map(d => d.file))];
    for (const file of files) {
        const saved = path.join(root,'.verbatim',file);
        fs.mkdirSync(path.dirname(saved),{recursive:true}); fs.copyFileSync(path.join(root,file),saved);
    }
    const applied = [];
    for (const number of numbers) {
        const directories = fs.readdirSync(path.join(__dirname,'adapt')).filter(n => n.startsWith(number+'-temporary-'));
        if (directories.length !== 1) throw new Error('missing or ambiguous adaptation '+number);
        execFileSync(process.execPath,[path.join(__dirname,'adapt',directories[0],'adapt.cjs'),root],{stdio:'inherit'});
        applied.push(directories[0]);
    }
    manifest.adaptations = {profile:'scanner', applied, final_sha256:Object.fromEntries(files.map(file => [file,crypto.createHash('sha256').update(fs.readFileSync(path.join(root,file))).digest('hex')]))};
    fs.writeFileSync(manifestPath,JSON.stringify(manifest,null,2)+'\n');
    execFileSync(process.execPath,[path.join(__dirname,'verify.cjs'),root],{stdio:'inherit'});
}
module.exports = {apply};
if (require.main === module) apply(process.argv[2]);
