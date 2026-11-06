"use strict";
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict");
const sites = require("./maybebind-sites.json");
// Only the twelve expressions in the two already pinned owners are candidates.
// Owner runtime hashes reject partial states before this planner can propose edits.
function plan(ts, tree, check = false) {
    const options = {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.NodeNext, moduleResolution: ts.ModuleResolutionKind.NodeNext, noEmit: true, skipLibCheck: true};
    const roots = ts.sys.readDirectory(path.join(tree, "src/compiler"), [".ts"], undefined, ["**/*.ts"]);
    const program = ts.createProgram(roots, options), checker = program.getTypeChecker();
    const source = program.getSourceFile(path.join(tree, "src/compiler/watch.ts"));
    const core = program.getSourceFile(path.join(tree, "src/compiler/core.ts"));
    assert(source && core, "maybeBind program sources missing");
    assert.equal(source.parseDiagnostics.length, 0, "maybeBind watch parse error");
    const helpers = core.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === "maybeBind");
    assert.equal(helpers.length, 1, "core maybeBind owner drift");
    function symbol(n) {
        let s = checker.getSymbolAtLocation(n);
        const seen = new Set();
        while (s && s.flags & ts.SymbolFlags.Alias) {
            assert(!seen.has(s), "maybeBind alias cycle"); seen.add(s);
            s = checker.getAliasedSymbol(s);
        }
        return s;
    }
    const target = symbol(helpers[0].name);
    assert(target, "core maybeBind symbol missing");
    const edits = [], states = [];
    for (const owner of [...new Set(sites.map(s => s.owner))]) {
        const owners = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === owner && n.body);
        assert.equal(owners.length, 1, "maybeBind owner drift: " + owner);
        const rows = sites.filter(s => s.owner === owner), found = new Map();
        function visit(n) {
            if (ts.isCallExpression(n) && ts.isIdentifier(n.expression) && (symbol(n.expression) === target || n.expression.text === "maybeBind")) {
                assert.equal(symbol(n.expression), target, "maybeBind callee symbol drift: " + owner);
                assert(!n.questionDotToken && !n.typeArguments && n.arguments.length === 2, "maybeBind call shape drift: " + owner);
                const [x, m] = n.arguments;
                assert(ts.isIdentifier(x) && ts.isPropertyAccessExpression(m) && !m.questionDotToken && ts.isIdentifier(m.expression), "maybeBind receiver shape drift: " + owner);
                assert(symbol(x) && symbol(x) === symbol(m.expression), "maybeBind receiver symbol drift: " + owner);
                const row = rows.find(s => s.receiver === x.text && s.member === m.name.text);
                assert(row && n.getText(source) === row.before && !found.has(row), "unlisted maybeBind site: " + owner);
                found.set(row, "before");
                edits.push({at: n.getStart(source), end: n.end, text: row.after});
            }
            if (ts.isCallExpression(n) && ts.isPropertyAccessExpression(n.expression) && n.expression.name.text === "bind") {
                const m = n.expression.expression;
                const row = ts.isPropertyAccessExpression(m) && ts.isIdentifier(m.expression) && rows.find(s => s.receiver === m.expression.text && s.member === m.name.text);
                if (row) {
                    assert(n.expression.questionDotToken && !m.questionDotToken && !n.questionDotToken && !n.typeArguments && n.arguments.length === 1 && ts.isIdentifier(n.arguments[0]), "optional bind shape drift: " + owner);
                    assert(symbol(m.expression) && symbol(m.expression) === symbol(n.arguments[0]), "optional bind symbol drift: " + owner);
                    assert(n.getText(source) === row.after && !found.has(row), "optional bind site drift: " + owner);
                    found.set(row, "after");
                }
            }
            ts.forEachChild(n, visit);
        }
        visit(owners[0]);
        assert.equal(found.size, rows.length, "maybeBind site inventory drift: " + owner);
        assert.equal(new Set(found.values()).size, 1, "half-adapted maybeBind owner: " + owner);
        states.push({owner, state: [...found.values()][0], sites: found.size});
    }
    if (check && edits.length) throw new Error("maybeBind rewrites missing");
    let text = source.text;
    for (const e of edits.sort((a,b) => b.at - a.at)) text = text.slice(0,e.at) + e.text + text.slice(e.end);
    return {text, rewrites: edits.length, states, javascriptChanging: true};
}
module.exports = {plan, sites, javascriptChanging: true};
