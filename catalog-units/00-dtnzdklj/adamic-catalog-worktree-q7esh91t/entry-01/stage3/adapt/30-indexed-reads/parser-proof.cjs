"use strict";
const assert = require("node:assert/strict");
function normalizeMissingJS(text) {
    const after = '        const missingNode = nodeIsMissing(node);\n        if (missingNode || node === undefined || intersectsIncrementalChange(node) || containsParseError(node)) {';
    const before = '        if (nodeIsMissing(node) || intersectsIncrementalChange(node) || containsParseError(node)) {';
    assert(text.split(after).length <= 2, "duplicate parser missing-node adaptation");
    return text.replace(after, before);
}
function compareMissingGuards(ts, before, after, utilities) {
    function guard(text) {
        const source = ts.createSourceFile("parser.ts", text, ts.ScriptTarget.Latest, true), found = [];
        function visit(node) {
            if (ts.isFunctionDeclaration(node) && node.name.text === "currentNode" && node.parameters[0]?.name.getText(source) === "parsingContext") found.push(node);
            ts.forEachChild(node, visit);
        }
        visit(source); assert.equal(found.length, 1);
        const fn = found[0];
        const condition = fn.body.statements.find(node => ts.isIfStatement(node) && node.expression.getText(source).includes("intersectsIncrementalChange"));
        assert(condition);
        const cache = fn.body.statements.find(node => ts.isVariableStatement(node) && node.getText(source).includes("const missingNode"));
        return new Function("node", "nodeIsMissing", "intersectsIncrementalChange", "containsParseError", (cache ? cache.getText(source) + "\n" : "") + "return (" + condition.expression.getText(source) + ");");
    }
    const source = ts.createSourceFile("utilities.ts", utilities, ts.ScriptTarget.Latest, true);
    const helper = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === "nodeIsMissing");
    assert(helper && helper.body.statements[0].getText(source).replace(/\s+/g," ") === "if (node === undefined) { return true; }");
    const js = ts.transpileModule(helper.getText(source), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2024 } }).outputText;
    const nodeIsMissing = new Function("SyntaxKind", "exports", js + "; return exports.nodeIsMissing;")(ts.SyntaxKind, {});
    const original = guard(before), adapted = guard(after);
    let cases = 0;
    function observe(fn, value, intersects, errors) {
        const trace = [];
        const node = value === undefined ? undefined : new Proxy(value, { get(target,key,receiver) { trace.push("read:" + String(key)); return Reflect.get(target,key,receiver); } });
        const result = fn(node, value => { trace.push("nodeIsMissing"); return nodeIsMissing(value); }, value => { trace.push("intersects"); assert(value !== undefined); return intersects; }, value => { trace.push("containsParseError"); assert(value !== undefined); return errors; });
        return { result, trace };
    }
    const nodes = [undefined];
    for (const pos of [-1,0,1,NaN]) for (const end of [-1,0,1,2,NaN]) for (const kind of [ts.SyntaxKind.EndOfFileToken,ts.SyntaxKind.Identifier]) nodes.push({pos,end,kind});
    for (const node of nodes) for (const intersects of [false,true]) for (const errors of [false,true]) {
        assert.deepEqual(observe(adapted,node,intersects,errors),observe(original,node,intersects,errors),"parser guard result or reads changed");cases++;
    }
    return { parser_missing_guard_node_cases: cases, results_and_calls_and_property_read_traces: "identical" };
}
module.exports = { normalizeMissingJS, compareMissingGuards };
