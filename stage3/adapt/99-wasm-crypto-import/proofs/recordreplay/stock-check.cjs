'use strict';
// Stock TypeScript with real @types/node and the real adapted declaration.
// Do not use plannedProgram/virtualHost or inject Adamic's Node augmentation.
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('typescript');
const {compiler, nodes} = require('../../../47-host-errors/adapt.cjs');
const recordreplay = require('../../recordreplay.cjs');
function check(tree) {
    tree = path.resolve(tree);
    const base = compiler(tree, ts), roots = base.getRootFileNames(), options = base.getCompilerOptions();
    const fileName = path.join(tree, 'src/compiler/sys.ts'), declarationName = path.join(tree, 'src/recordreplay.d.ts');
    assert.equal(roots.length, 79);
    assert.equal(fs.readFileSync(declarationName, 'utf8'), recordreplay.declaration);
    const nodePackage = require.resolve('@types/node/package.json');
    assert.equal(require(nodePackage).version, require('../../../../api/package.json').dependencies['@types/node']);
    function program(withoutDeclaration) {
        const host = ts.createCompilerHost(options, true);
        host.resolveTypeReferenceDirectives = (names, containingFile, redirectedReference, opts, containingFileMode) => names.map(name =>
            ts.resolveTypeReferenceDirective(typeof name === 'string' ? name : name.fileName, containingFile,
                {...opts, typeRoots:[path.dirname(path.dirname(nodePackage))]}, host, redirectedReference, undefined, containingFileMode).resolvedTypeReferenceDirective);
        if (withoutDeclaration) {
            const exists = host.fileExists.bind(host), read = host.readFile.bind(host), source = host.getSourceFile.bind(host);
            host.fileExists = name => path.resolve(name) !== declarationName && exists(name);
            host.readFile = name => path.resolve(name) === declarationName ? undefined : read(name);
            host.getSourceFile = (name, ...args) => path.resolve(name) === declarationName ? undefined : source(name, ...args);
        }
        return ts.createProgram(roots, options, host);
    }
    function diagnostics(p) { return ts.getPreEmitDiagnostics(p).map(d => ({code:d.code, file:d.file?.fileName, start:d.start, message:ts.flattenDiagnosticMessageText(d.messageText, ' ')})); }
    const positive = program(false), file = positive.getSourceFile(fileName);
    const owner = nodes(ts, file, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem');
    assert.equal(owner.length, 1); recordreplay.validate(positive, file, owner[0]);
    const withDeclaration = diagnostics(positive);
    assert.equal(withDeclaration.length, 0, JSON.stringify(withDeclaration));
    const negative = program(true), withoutDeclaration = diagnostics(negative);
    const missingMember = withoutDeclaration.filter(d => d.code === 2339 && d.file === fileName && d.message.includes('recordreplay'));
    assert(missingMember.length > 0, 'stock tsc without the declaration must report TS2339 on process.recordreplay');
    return {version:ts.version, nodeVersion:require(nodePackage).version, roots, injectedAugmentation:false, withDeclaration, withoutDeclaration};
}
module.exports = {check};
if (require.main === module) {
    const result = check(process.argv[2]);
    if (process.argv[3]) fs.writeFileSync(process.argv[3], JSON.stringify(result,null,2)+'\n');
    console.log(JSON.stringify(result,null,2));
}
