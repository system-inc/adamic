'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const [tree, output] = process.argv.slice(2).map(p => path.resolve(p));
const names = ['addNodeFactoryPatcher', 'addObjectAllocatorPatcher'];
const walk = directory => fs.readdirSync(directory, {withFileTypes:true}).flatMap(e => e.isDirectory() ? walk(path.join(directory,e.name)) : [path.join(directory,e.name)]);
// Parse all source, harness, test inputs and baseline programs, including JavaScript.
// A syntax inventory selects test candidates; actual references use the language service.
const all = ['src','tests','scripts'].flatMap(d => walk(path.join(tree,d))).filter(f => /\.(?:[cm]?[jt]s|[jt]sx)$/.test(f));
const candidates = [], occurrences = [];
for (const file of all) {
    const source = ts.createSourceFile(file, fs.readFileSync(file,'utf8'), ts.ScriptTarget.Latest, true);
    let found = false;
    const pending = [source];
    while (pending.length) {
        const n = pending.pop();
        if ((ts.isIdentifier(n) || ts.isStringLiteral(n)) && names.includes(n.text)) {
            found = true;
            occurrences.push({file:path.relative(tree,file), name:n.text, kind:ts.SyntaxKind[n.kind], parent:ts.SyntaxKind[n.parent.kind], start:n.getStart(source)});
        }
        ts.forEachChild(n, child => { pending.push(child); });
    }
    if (found) candidates.push(file);
}
const roots = [...new Set([...all.filter(f => f.startsWith(path.join(tree,'src')+path.sep)), ...candidates])];
const options = {target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,allowJs:true,strict:true,noEmit:true,skipLibCheck:true};
const host = {getScriptFileNames:()=>roots,getScriptVersion:()=> '0',getScriptSnapshot:f=>{const t=ts.sys.readFile(f);return t===undefined?undefined:ts.ScriptSnapshot.fromString(t);},getCurrentDirectory:()=>tree,getCompilationSettings:()=>options,getDefaultLibFileName:o=>ts.getDefaultLibFilePath(o),fileExists:ts.sys.fileExists,readFile:ts.sys.readFile,readDirectory:ts.sys.readDirectory,directoryExists:ts.sys.directoryExists,getDirectories:ts.sys.getDirectories};
const service = ts.createLanguageService(host);
const program = service.getProgram();
const rows = [];
for (const [index,name] of names.entries()) {
    const relative = index === 0 ? 'src/compiler/factory/nodeFactory.ts' : 'src/compiler/utilities.ts';
    const file = path.join(tree,relative), source = program.getSourceFile(file);
    const declaration = source.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === name);
    assert(declaration, name);
    const groups = service.findReferences(file, declaration.name.getStart(source));
    assert(groups?.length, 'findReferences returned no definition');
    const references = groups.flatMap(g => g.references).map(r => {
        const sf = program.getSourceFile(r.fileName);
        function at(n) { if (n.getStart(sf) <= r.textSpan.start && n.end >= r.textSpan.start+r.textSpan.length) { for (const c of n.getChildren(sf)) { const found=at(c); if(found)return found; } return n; } }
        const node = at(sf), parent = node.parent;
        const permitted = (ts.isFunctionDeclaration(parent) && parent.name === node) || ts.isExportSpecifier(parent) || ts.isImportSpecifier(parent);
        const position=sf.getLineAndCharacterOfPosition(r.textSpan.start);
        return {file:path.relative(tree,r.fileName),line:position.line+1,column:position.character+1,parent:ts.SyntaxKind[parent.kind],text:parent.getText(sf),permitted};
    });
    rows.push({name,references,callers:references.filter(r=>!r.permitted)});
}
const report={typescript:ts.version,method:'AST inventory plus LanguageService.findReferences',filesParsed:all.length,sourceRoots:roots.length,occurrences,registrations:rows};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');
assert(rows.every(r=>r.callers.length===0),'registrant or value alias exists');
console.log(JSON.stringify({filesParsed:all.length,sourceRoots:roots.length,registrations:rows}));
