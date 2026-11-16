from pathlib import Path
import subprocess as s
R=Path('/workspace/scratch/b7phwab-amend7/work');D=R/'stage3/adapt/99-wasm-crypto-import'
p=D/'recordreplay.cjs';old=p.read_text();original=s.check_output(['git','show','ff0e5859:stage3/adapt/99-wasm-crypto-import/recordreplay.cjs'],cwd=R,text=True)
# The original declaration-bearing helper already has the exact AST lookup.
assert old[old.index('function accesses'):old.index('function edits')].rstrip()==original[original.index('function accesses'):original.index('function sourceState')].rstrip()
original=original.replace('// The compiler owns this optional absent member already. This declaration lets\n// stock tsc read the same member without a cast; it promises no runtime value.','// Stock tsc builds the adapted compiler with @types/node, without Adamic\'s\n// Node prelude. This erased declaration is required by that hereby local build.')
p.write_text(original)
p=D/'adapt.cjs';x=p.read_text();x=x.replace('    let importInsertions = 0;','    let importInsertions = 0;\n    let recordreplayImports = 0;').replace("        assert(inOwner || importInsertion, 'edit outside getNodeSystem and pinned import edits');","        const recordreplayImport = recordreplay.importEdit(osImport, newline, edit);\n        assert(inOwner || importInsertion || recordreplayImport, 'edit outside getNodeSystem and pinned import edits');\n        if (recordreplayImport) recordreplayImports++;").replace("    assert.equal(importInsertions, done ? 0 : 1, 'crypto import insertion count moved');","    assert.equal(importInsertions, done ? 0 : 1, 'crypto import insertion count moved');\n    assert.equal(recordreplayImports, done ? 0 : 1, 'recordreplay import edit count moved');").replace("    if (edits.length) fs.writeFileSync(file.fileName, text);","    recordreplay.write(file, done);\n    if (edits.length) fs.writeFileSync(file.fileName, text);").replace('files: edits.length ? 1 : 0','files: edits.length ? 2 : 0');p.write_text(x)
p=D/'proofs/check.cjs';x=p.read_text().replace("const watchDirectory = require('./watch-directory.json');","const watchDirectory = require('./watch-directory.json');\nconst recordreplay = require('../recordreplay.cjs');").replace("    if (!erased) {\n        text = replaceOne(text, 'process.recordreplay'","    if (!erased) {\n        text = replaceOne(text, recordreplay.importText, '');\n        text = replaceOne(text, 'process.recordreplay'")
x=x.replace("    assert.deepEqual(afterNames, names, 'source inventory changed');","    const added = 'src/recordreplay.d.ts';\n    assert(!names.includes(added));\n    assert.deepEqual(afterNames, [...names, added].sort(), 'source inventory changed outside the restored declaration');\n    assert.equal(fs.readFileSync(path.join(after, added), 'utf8'), recordreplay.declaration);").replace("    assert(!afterNames.includes('src/recordreplay.d.ts'), 'adaptation declaration remains');\n",'').replace("    const changed = entries.filter", "    entries.push({file:added,before:null,after:hash(fs.readFileSync(path.join(after,added)))});\n    const changed = entries.filter").replace("['src/compiler/sys.ts'], 'source changed outside sys.ts'","['src/compiler/sys.ts', added], 'source changed outside sys.ts and the restored declaration'").replace('added:[], changed','added:[added], changed');p.write_text(x)
p=D/'proofs/recordreplay-contract.cjs';x=p.read_text().replace('adapt(scratch).files, 1','adapt(scratch).files, 2');a=x.index("    assert(!fs.existsSync(declarationName)");b=x.index('    // Relocate',a);x=x[:a]+'''    assert.equal(fs.readFileSync(declarationName, 'utf8'), recordreplay.declaration);
    const stamps = [fileName, declarationName].map(p => fs.statSync(p, {bigint: true}).mtimeNs);
    assert.equal(adapt(scratch).files, 0);
    assert.deepEqual([fileName, declarationName].map(p => fs.statSync(p, {bigint: true}).mtimeNs), stamps);
    const stock = require('./recordreplay/stock-check.cjs').check(scratch);
    assert.equal(stock.withDeclaration.length, 0);
    assert(stock.withoutDeclaration.some(d => d.code === 2339 && d.message.includes('recordreplay')));
'''+x[b:];x=x.replace('adaptationDeclarationAbsent: true,augmentationReadTypes: true,withoutAugmentationTS2339: true','stockBuildDeclarationPresent: true,stockReadTypes: true,withoutDeclarationTS2339: true');p.write_text(x)
# These existing contracts count the adapter's writes and seed a pre-99 source.
p=D/'proofs/require-result-contract.cjs';p.write_text(p.read_text().replace('adapt(scratch).files, 1','adapt(scratch).files, 2'))
p=D/'proofs/watch-directory-contract.cjs';x=p.read_text().replace('adapter.adapt(root).files,1','adapter.adapt(root).files,2').replace("const original=before.replace", "fs.rmSync(path.join(root,'src/recordreplay.d.ts'),{force:true});\n  const original=before.replace");p.write_text(x)
p=D/'README.md';x=p.read_text();a=x.index('## Recordreplay amendment, task #b7phwab');x=x[:a]+'''## Recordreplay amendment, task #b7phwab

The adapter removes `as any` and its parentheses from getNodeSystem's
nodeSystem.debugMode read of `process.recordreplay`. It restores exactly
ff0e5859's `src/recordreplay.d.ts` bytes and the erased
`import type {} from "../recordreplay.js";` in the reserved import line.
The AST lookup and all composed adapter-99 edits from 163bdeaf remain intact.
The declaration and import are outside getNodeSystem and the pinned node:
import block, so restoring them does not change an owner pin.

The file is required for **stock tsc's build of the adapted tree**:
stage3/lane/adaptedtree's TestProduct_Stage3AdaptedTree runs hereby local with
@types/node. Adamic's Node-prelude augmentation is visible only to Adamic's
loader; that stock build never sees it. The file was dropped because the two
environments were confused, and is restored because hereby local reports
TS2339 without it. The declaration adds no runtime value; its import erases.

`proofs/recordreplay/check.py SCRATCH` checks exact source scope, canonical
file bytes and no-write replay, then runs the actual Adamic loader and a
separate unmodified stock TypeScript program with @types/node. The stock
program loads the real file through the real import: it injects no Node
augmentation or virtual declaration, and masking only that file must give
TS2339 for recordreplay. `ADAMIC_BIN` names the production CLI for the loader
check. `proofs/recordreplay-contract.cjs` retains AST relocation, wrong-store,
wrong-property, duplicate-read and shadow rejection controls.

The acceptance build is TestProduct_Stage3AdaptedTree itself, including hereby
local and hereby tests. Its negative control removes the declaration after
apply and before hereby local, so the same product test must fail with TS2339.
Historical amendment-5 evidence used an injected stock prelude and did not
prove this build; it is not the acceptance proof for this restoration.
''';p.write_text(x)
