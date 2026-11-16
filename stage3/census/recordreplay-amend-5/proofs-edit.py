from pathlib import Path
D=Path('/workspace/scratch/b7phwab-amend5/work/stage3/adapt/99-wasm-crypto-import')
p=D/'proofs/recordreplay-contract.cjs';s=p.read_text().replace('adapt(scratch).files, 2','adapt(scratch).files, 1');a=s.index("    assert.equal(fs.readFileSync(declarationName");b=s.index('    // Relocate',a);s=s[:a]+'''    assert(!fs.existsSync(declarationName), 'adapter must not write a declaration');
    const stamp = fs.statSync(fileName, {bigint: true}).mtimeNs;
    assert.equal(adapt(scratch).files, 0);
    assert.equal(fs.statSync(fileName, {bigint: true}).mtimeNs, stamp);
    const augmented = plannedProgram(compiler(scratch, ts), parse(expected), expected);
    const af = augmented.getSourceFile(fileName);
    recordreplay.validate(augmented, af, owner(af));
    const read = recordreplay.accesses(af, owner(af));
    assert(!augmented.getSemanticDiagnostics(af).some(d => d.start <= read.end && d.start + d.length >= read.pos), 'augmented read has a diagnostic');
    const original = recordreplay.virtualHost;
    recordreplay.virtualHost = host => {
        const nodePackage = require.resolve('@types/node/package.json');
        host.resolveTypeReferenceDirectives = (names, containingFile, redirectedReference, options, containingFileMode) => names.map(name => ts.resolveTypeReferenceDirective(typeof name === 'string' ? name : name.fileName, containingFile, {...options, typeRoots:[path.dirname(path.dirname(nodePackage))]}, host, redirectedReference, undefined, containingFileMode).resolvedTypeReferenceDirective);
    };
    let control;
    try { control = plannedProgram(compiler(scratch, ts), af, expected); }
    finally { recordreplay.virtualHost = original; }
    assert(control.getSemanticDiagnostics(control.getSourceFile(fileName)).some(d => d.code === 2339 && ts.flattenDiagnosticMessageText(d.messageText, ' ').includes('recordreplay')), 'removing CLI augmentation must give TS2339');
'''+s[b:];s=s.replace('declarationUnchanged: true','adaptationDeclarationAbsent: true,augmentationReadTypes: true,withoutAugmentationTS2339: true');s=s.replace('plannedProgram(compiler(scratch, ts), parse(expected), expected)','plannedProgram(compiler(scratch, ts), {fileName}, expected)');p.write_text(s)
p=D/'proofs/require-result-contract.cjs';s=p.read_text().replace('adapt(scratch).files, 2','adapt(scratch).files, 1');p.write_text(s)
p=D/'proofs/scoped.py';s=p.read_text().replace("b'import * as _crypto from \"node:crypto\";'+newline+b'import type {} from \"../recordreplay.js\";'+newline","b'import * as _crypto from \"node:crypto\";'+newline");s='\n'.join(l for l in s.split('\n') if "expectedSources['src/recordreplay.d.ts']" not in l);s=s.replace("['files']==2","['files']==1").replace('inOwner || importInsertion || recordreplayImport','inOwner || importInsertion');p.write_text(s)
p=D/'README.md';s=p.read_text();a=s.index('## Recordreplay amendment, task #b7phwab');s=s[:a]+'''## Recordreplay amendment, task #b7phwab

Re-cut on main `34c05e3a4f79fa64ef91e9200df03af4b0c02208` with CLI's
three commits on the same parent. All landed amendments above remain composed.
The adapter removes only `as any` and its parentheses from the
getNodeSystem nodeSystem.debugMode read of `process.recordreplay`.
It adds no source file or import: the previous `src/recordreplay.d.ts` and
`import type {} from "../recordreplay.js"` are dropped because CLI's
`internal/load/node_library.go` Node-prelude augmentation declares the member.

The validation host reads that exact augmentation from CLI and appends it to
the pinned Node declarations in memory, resolving the API's locked @types/node
without requiring upstream node_modules. The direct read binds to ambient
NodeJS.Process and has type unknown. The contract checks that removing only
the augmentation gives TS2339 for recordreplay. No adapter declaration is
written, and replay performs no source write. Source inventory stays unchanged.

Owner pins are regenerated through scoped-hashes.cjs and adapt.cjs --pin
against composed apply output. Fresh proof results are recorded separately on
typescript/recordreplay-amend-5-evidence. Historical recordreplay inventory and
result JSON files describe earlier declaration-bearing runs, not this re-cut.
No Loom submission is made.
''';p.write_text(s)
