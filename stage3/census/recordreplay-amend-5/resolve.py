from pathlib import Path
import re,subprocess
root=Path('/workspace/scratch/b7phwab-amend5/work'); d=root/'stage3/adapt/99-wasm-crypto-import'
def sides(p,which):
 s=p.read_text(); s=re.sub(r'^<<<<<<< HEAD\n(.*?)^=======\n(.*?)^>>>>>>>[^\n]*\n',lambda m:m[1] if which=='ours' else m[2] if which=='theirs' else m[1]+m[2],s,flags=re.M|re.S);p.write_text(s)
sides(d/'README.md','both');sides(d/'adapt.cjs','ours');sides(d/'proofs/check.cjs','both');sides(d/'proofs/scope-mutant.cjs','theirs')
(d/'proofs/targets.json').write_bytes(subprocess.check_output(['git','show','HEAD:stage3/adapt/99-wasm-crypto-import/proofs/targets.json'],cwd=root))
p=d/'adapt.cjs';s=p.read_text().replace('    let recordreplayImports = 0;\n','').replace('        const recordreplayImport = recordreplay.importEdit(osImport, newline, edit);\n','').replace('inOwner || importInsertion || recordreplayImport','inOwner || importInsertion').replace('        if (recordreplayImport) recordreplayImports++;\n','').replace("    assert.equal(recordreplayImports, done ? 0 : 1, 'recordreplay import edit count moved');\n",'').replace('files: edits.length ? 2 : 0','files: edits.length ? 1 : 0');p.write_text(s)
p=d/'recordreplay.cjs';s=p.read_text();start=s.index('const importText');end=s.index('function accesses');s=s[:start]+'''// Read CLI's augmentation verbatim; validation must follow the compiler's prelude.
const loader = fs.readFileSync(path.resolve(__dirname, '../../../internal/load/node_library.go'), 'utf8');
const augmentation = JSON.parse(loader.match(/^const nodeProcessAugmentation = (".*")$/m)[1]);
'''+s[end:];start=s.index('function sourceState');end=s.index('function virtualHost');s=s[:start]+'''function edits(file, owner, os, newline, done) {
    assert(!fs.existsSync(path.resolve(path.dirname(file.fileName), '../recordreplay.d.ts')), 'adaptation declaration must be absent');
    assert(!file.statements.some(n => ts.isImportDeclaration(n) && n.moduleSpecifier.text === '../recordreplay.js'), 'adaptation import must be absent');
    const read = accesses(file, owner);
    assert.equal(read.getText(file), done ? 'process.recordreplay' : '(process as any).recordreplay', 'recordreplay receiver changed');
    return done ? [] : [{start: read.getStart(file), end: read.end, text: 'process.recordreplay'}];
}
'''+s[end:];start=s.index('    const target = declarationPath(file);');end=s.index('\nfunction validate',start);s=s[:start]+'''    const source = host.getSourceFile.bind(host);
    host.getSourceFile = (name, language, onError, fresh) => {
        const result = source(name, language, onError, fresh);
        if (result && path.resolve(name) === path.join(path.dirname(nodePackage), 'globals.d.ts'))
            return ts.createSourceFile(name, result.text + augmentation, language, true);
        return result;
    };
}
'''+s[end:];start=s.index('    const imported =',s.index('function validate'));end=s.index('    const read =',start);s=s[:start]+s[end:];start=s.index('    const declarations = program.getSourceFiles()');s=s[:start]+'''    const member = checker.getPropertyOfType(checker.getTypeAtLocation(read.expression), 'recordreplay');
    assert(member?.declarations?.some(d => d.getSourceFile().text.endsWith(augmentation)), 'read must type through CLI Node-prelude augmentation');
    return {directRead: true, type: 'unknown', nodePreludeAugmentation: true, adaptationDeclaration: false};
}
module.exports = {accesses, edits, virtualHost, validate, augmentation};
''';p.write_text(s)
p=d/'proofs/check.cjs';s=p.read_text().replace("        text = replaceOne(text, recordreplay.importText, '');\n",'');start=s.index("    const added = 'src/recordreplay.d.ts';");end=s.index('    const a = fs.readFileSync',start);s=s[:start]+'''    assert.deepEqual(afterNames, names, 'source inventory changed');
    assert(!afterNames.includes('src/recordreplay.d.ts'), 'adaptation declaration remains');
    const entries = names.map(file => ({file, before:hash(fs.readFileSync(path.join(before,file))), after:hash(fs.readFileSync(path.join(after,file)))}));
    const changed = entries.filter(x => x.before !== x.after).map(x => x.file).sort();
    assert.deepEqual(changed, ['src/compiler/sys.ts'], 'source changed outside sys.ts');
'''+s[end:];s=s.replace('added:[added]','added:[]');p.write_text(s)
