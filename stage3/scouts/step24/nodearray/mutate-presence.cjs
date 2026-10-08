// Make one real initializer mutation in a new scratch copy of the instrumented compiler.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const [source, target] = process.argv.slice(2).map(x=>path.resolve(x));
assert.equal(ts.version,'6.0.3');
assert.ok(!fs.existsSync(target),'mutation destination must be new');
const name='src/compiler/factory/nodeFactory.ts';
const text=fs.readFileSync(path.join(source,name),'utf8');
const tree=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);
const selections=[];
function walk(node) {
    if(ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression)
        && node.expression.name.text==='__nodearrayWrite'
        && node.arguments[1]?.text==='hasTrailingComma'
        && node.arguments[3]?.text==='src/compiler/factory/nodeFactory.ts:1202:9') selections.push(node.arguments[2]);
    ts.forEachChild(node,walk);
}
walk(tree);
assert.equal(selections.length,1);
const value=selections[0];
assert.ok(ts.isPrefixUnaryExpression(value)&&value.operator===ts.SyntaxKind.ExclamationToken);
assert.ok(ts.isPrefixUnaryExpression(value.operand)&&value.operand.operator===ts.SyntaxKind.ExclamationToken);
assert.equal(value.operand.operand.getText(tree),'hasTrailingComma');
const replacement=value.operand.operand.getText(tree);
fs.cpSync(source,target,{recursive:true});
fs.writeFileSync(path.join(target,name),text.slice(0,value.getStart(tree))+replacement+text.slice(value.end));
const record={file:name+':1202',before:value.getText(tree),after:replacement,expected:'present-undefined hasTrailingComma appears and semantic digest changes'};
fs.writeFileSync(path.join(target,'presence-mutant.json'),JSON.stringify(record,null,2)+'\n');
console.log(JSON.stringify(record));
