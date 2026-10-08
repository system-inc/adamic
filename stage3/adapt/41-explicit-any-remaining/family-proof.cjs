"use strict";
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict"), crypto = require("node:crypto"), ts = require("typescript");
const [beforeTree, afterTree, family, output] = process.argv.slice(2);
const rules = require("./rules.json").filter(r => r.kind === family);
assert(rules.length, "reviewed family");
const hash = s => crypto.createHash("sha256").update(s).digest("hex");
const normalize = node => {
    if (ts.isParenthesizedExpression(node)) return normalize(node.expression);
    const children = []; node.forEachChild(child => { children.push(normalize(child)); });
    return [node.kind, children.length ? children : node.getText()];
};
const emit = (text, file) => ts.transpileModule(text, { fileName: file, compilerOptions: { target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext } }).outputText;
const inventory = text => {
    let count = 0; const visit = n => { if (n.kind === ts.SyntaxKind.AnyKeyword) count++; n.forEachChild(visit); };
    visit(ts.createSourceFile("source.ts", text, ts.ScriptTarget.Latest, true)); return count;
};
const evidence = { family, typescript: ts.version, files: {}, tokensRemoved: 0, mutants: [] };
for (const file of [...new Set(rules.map(r => r.file))]) {
    const before = fs.readFileSync(path.join(beforeTree,file),"utf8"), after = fs.readFileSync(path.join(afterTree,file),"utf8");
    let expected = before;
    for (const rule of rules.filter(r => r.file === file)) {
        assert.equal(expected.split(rule.before).length,2,"one original owner " + rule.id);
        expected = expected.replace(rule.before,rule.after);
    }
    assert.equal(after,expected,"only reviewed family edits " + file);
    const a = emit(before,file), b = emit(after,file);
    const ast = s => normalize(ts.createSourceFile("emitted.js",s,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS));
    assert.deepEqual(ast(b),ast(a),"identical JavaScript AST modulo parentheses " + file);
    const removed = inventory(before)-inventory(after); assert(removed >= 0);
    evidence.tokensRemoved += removed;
    evidence.files[file] = { before:hash(before), after:hash(after), tokensRemoved:removed, javascriptByteIdentical:a===b, javascriptASTIdentical:true };
    const realMutation = b + "\nthrow new Error('family emission mutant');\n";
    assert.throws(() => assert.deepEqual(ast(realMutation),ast(a)),assert.AssertionError);
    evidence.mutants.push({file, mutation:"add throw to actual emitted source", caughtBy:"whole-file JavaScript AST equality"});
}
fs.writeFileSync(output,JSON.stringify(evidence,null,2)+"\n");
console.log(JSON.stringify(evidence));
