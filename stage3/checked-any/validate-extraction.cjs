// Verify unchanged stock declarations and declaration-only surrounding context.
const fs = require('fs');
const crypto = require('crypto');
const ts = require('../api/node_modules/typescript');
const [stock, manifestPath] = process.argv.slice(2);
const records = JSON.parse(fs.readFileSync(manifestPath));
function ambient(node) {
    if (ts.isTypeAliasDeclaration(node) || ts.isInterfaceDeclaration(node) || ts.isEnumDeclaration(node) || ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) return true;
    if (ts.isVariableStatement(node)) return node.modifiers?.some(m=>m.kind===ts.SyntaxKind.DeclareKeyword) && node.declarationList.declarations.every(d=>!d.initializer);
    if (ts.isModuleDeclaration(node) && node.body && ts.isModuleBlock(node.body)) return node.body.statements.every(ambient);
    return false;
}
for (const record of records) {
    const original = fs.readFileSync(stock+'/'+record.file, 'utf8').slice(record.declaration_start, record.declaration_end);
    const source = fs.readFileSync(record.source, 'utf8');
    if (crypto.createHash('sha256').update(original).digest('hex') !== record.declaration_sha256 || source.slice(record.declaration_start, record.declaration_end) !== original) throw Error('stock declaration text changed at record '+record.index);
    const parsed = ts.createSourceFile(record.source, source, ts.ScriptTarget.Latest, true);
    if (parsed.parseDiagnostics.length) throw Error('invalid extracted syntax at record '+record.index);
    for (const node of parsed.statements) {
        if (node.getStart(parsed)>=record.declaration_start && node.end<=record.declaration_end) continue;
        if (!ambient(node)) throw Error('substituted runtime body outside record '+record.index);
    }
}
console.log('verified '+records.length+' unchanged declarations and declaration-only contexts');
