// Identify the shared stock return token with the compiler API, then audit its ledger.
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const [prepared, result] = process.argv.slice(2);
const file = 'src/compiler/sys.ts';
const source = ts.createSourceFile(file, fs.readFileSync(path.join(prepared, 'stock', file), 'utf8'), ts.ScriptTarget.Latest, true);
const timer = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === 'setTimeout');
assert.equal(timer.length, 1);
assert.equal(timer[0].type.kind, ts.SyntaxKind.AnyKeyword);
const start = timer[0].type.getStart(source);
const ledger = JSON.parse(fs.readFileSync(path.join(result, 'ledger.json'), 'utf8'));
function audit(rows) {
    const sites = rows.filter(r => r.file === file && r.kind === 'explicit_any' && r.node_start === start);
    assert.equal(sites.length, 1, 'shared timer return counted more than once');
    assert.equal(sites[0].disposition, 'rewritten');
    assert.equal(sites[0].adaptation, '41-explicit-any-remaining');
    assert.equal(new Set(rows.map(r => r.id)).size, rows.length, 'duplicate stock site ID');
    return sites[0];
}
const site = audit(ledger);
const mutant = [...ledger, {...site, adaptation:'43-any-returns'}];
assert.throws(() => audit(mutant), /shared timer return counted more than once/);
const composition = JSON.parse(fs.readFileSync(path.join(prepared, 'composition.json'), 'utf8'));
assert.equal(composition.length, 3);
const timerComposition = composition.filter(c => c.skipped_rule === '43-any-returns:setTimeout');
assert.equal(timerComposition.length, 1);
const adapted = fs.readFileSync(path.join(prepared, 'adapted', file), 'utf8');
assert.equal(adapted.split(timerComposition[0].retained).length - 1, 1);
assert.equal(adapted.includes(timerComposition[0].skipped), false);
function auditContracts(text) {
    const file = ts.createSourceFile('commandLineParser.ts', text, ts.ScriptTarget.Latest, true);
    for (const [name, returnType] of [['convertConfigFileToObject','AdamicJsonRecoveryObject'], ['convertToJson','AdamicJsonRecoveryValue | undefined']]) {
        const functions = file.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === name);
        assert.equal(functions.length, 1);
        assert.equal(functions[0].type.getText(file), returnType, '43 return contract lost in composition');
        const errors = functions[0].parameters.filter(p => p.name.getText(file) === 'errors');
        assert.equal(errors.length, 1);
        assert.equal(errors[0].type.getText(file), 'Pick<DiagnosticWithLocation[], "push">');
    }
    return file;
}
const jsonText = fs.readFileSync(path.join(prepared, 'adapted/src/compiler/commandLineParser.ts'), 'utf8');
const jsonSource = auditContracts(jsonText);
const converter = jsonSource.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'convertToJson');
const lostContract = jsonText.slice(0, converter.type.getStart(jsonSource)) + 'any' + jsonText.slice(converter.type.end);
assert.throws(() => auditContracts(lostContract), /43 return contract lost in composition/);
console.log(JSON.stringify({stock_return_site:site.id,disposition:site.disposition,owner:site.adaptation,count:1,duplicate_credit_mutant_caught:true,json_contract_mutant_caught:true,composition}, null, 2));
