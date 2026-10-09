'use strict';
// Restore the measured pre-return-adaptation scanner corpus, with type imports only.
const fs = require('node:fs'), path = require('node:path'), ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('requires stock TypeScript 6.0.3');
const file = path.join(path.resolve(process.argv[2]),'src/compiler/scanner.ts');
const program = ts.createProgram([file],{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,allowImportingTsExtensions:true,verbatimModuleSyntax:true});
const source = program.getSourceFile(file), checker = program.getTypeChecker(), edits=[];
for (const node of source.statements) {
    if (!ts.isImportDeclaration(node) || node.importClause?.isTypeOnly || !ts.isNamedImports(node.importClause?.namedBindings)) continue;
    for (const item of node.importClause.namedBindings.elements) {
        if (item.isTypeOnly) continue;
        const alias = checker.getSymbolAtLocation(item.name);
        const symbol = alias && checker.getAliasedSymbol(alias);
        if (!symbol || symbol.flags === ts.SymbolFlags.Unknown) throw Error('unresolved corpus import: ' + item.name.text);
        if (!(symbol.flags & ts.SymbolFlags.Value)) edits.push(item.getStart(source));
    }
}
let text=source.text;
for (const position of edits.sort((a,b)=>b-a)) text=text.slice(0,position)+'type '+text.slice(position);
fs.writeFileSync(file,text);console.log(JSON.stringify({imports:edits.length}));
