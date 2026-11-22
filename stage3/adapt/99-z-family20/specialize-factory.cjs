const assert = require("node:assert/strict");
module.exports = function specializeFactory(ts, input, source) {
    const ast = ts.createSourceFile(input, source, ts.ScriptTarget.Latest, true);
    const eol = source.includes("\r\n") ? "\r\n" : "\n";
    const roles = new Map();
    for (const group of ["UnaryType", "PrePostfixUnaryType", "SimpleTag", "TypeLikeTag"]) {

        for (const action of ["Create", "Update"]) {
            roles.set("getJSDoc" + group + action + "Function",
                      action.toLowerCase() + "JSDoc" + group + "Worker");
        }
    }
    const workers = new Map([...roles.values()].map(name => [name, new Map()]));
    const bindings = new Map();
    const declarations = new Map();
    const uses = new Map([...roles.keys()].map(name => [name, new Map()]));
    const edits = [];
    const replace = (node, value) => edits.push({start: node.getStart(ast), end: node.end, value});
    function visit(node) {
        if (ts.isVariableDeclaration(node) && roles.has(node.name.getText(ast))) {
            bindings.set(node.name.getText(ast), node);
        }
        if (ts.isFunctionDeclaration(node) && workers.has(node.name?.text)) {
            assert(!declarations.has(node.name.text));
            declarations.set(node.name.text, node);
        }
        if (ts.isCallExpression(node) && uses.has(node.expression.getText(ast))) {
            const name = node.expression.getText(ast);
            assert.equal(node.typeArguments?.length, 1);
            assert.equal(node.arguments.length, 1);
            const type = node.typeArguments[0].getText(ast);
            assert.equal(node.arguments[0].getText(ast), "SyntaxKind." + type);
            assert(!uses.get(name).has(type));
            uses.get(name).set(type, node);
        }
        ts.forEachChild(node, visit);
    }
    visit(ast);

    function specialize(node, type) {
        const params = node.typeParameters;
        assert.equal(params?.length, 1);
        assert.equal(params[0].name.text, "T");
        const start = node.getStart(ast);
        const text = node.getText(ast);
        const open = text.lastIndexOf("<", params[0].getStart(ast) - start);
        const end = params[0].end - start;
        assert.equal(text[end], ">");
        return (text.slice(0, open) + text.slice(end + 1)).replace(/\bT\b/g, type);
    }

    for (const [name, worker] of roles) {
        const binding = bindings.get(name);
        const kindArrow = binding.initializer.arguments[0];
        assert(ts.isArrowFunction(kindArrow));
        assert.equal(binding.initializer.expression.getText(ast), "memoizeOne");
        const variants = [];
        for (const [type, call] of uses.get(name)) {
            const declarationName = name + "For" + type;
            const transform = specialize(kindArrow, type)
                .replace(new RegExp("\\b" + worker + "(?:<" + type + ">)?\\("), worker + "For" + type + "(");
            assert(transform.includes(worker + "For" + type));
            variants.push(declarationName + " = memoizeOne(" + transform + ")");
            replace(call, declarationName + "(" + call.arguments[0].getText(ast) + ")");
            workers.get(worker).set(type, true);
            if (worker.startsWith("update")) workers.get("create"+worker.slice(6)).set(type, true);
            if (worker.includes("PrePostfix")) workers.get("createJSDocUnaryTypeWorker").set(type, true);
        }
        assert(variants.length >= 2);
        replace(binding.parent.parent, variants.map(d => "const " + d + ";").join(eol + "    "));
    }

    for (const [name, types] of workers) {
        const base = declarations.get(name);
        const variants = [];
        for (const type of types.keys()) {
            let code = specialize(base, type);
            for (const worker of workers.keys()) {
                code = code.replace(new RegExp("\\b" + worker + "(?:<" + type + ">)?\\(", "g"),
                                    worker + "For" + type + "(");
            }
            if (workers.get("createJSDocPrePostfixUnaryTypeWorker").has(type)) {
                if (name === "createJSDocUnaryTypeWorker") {
                    const prior = "): " + type + " {";
                    assert(code.includes(prior));
                    code = code.replace(prior, "): Mutable<" + type + "> {");
                }
                if (name === "createJSDocPrePostfixUnaryTypeWorker") {
                    const cast = " as Mutable<" + type + ">";
                    assert(code.includes(cast));
                    code = code.replace(cast, "");
                }
            }
            variants.push(code);
        }
        replace(base, variants.join(eol + eol + "    "));
    }
    let result = source;
    for (const edit of edits.sort((a,b) => b.start-a.start)) {
        result = result.slice(0, edit.start) + edit.value + result.slice(edit.end);
    }
    const parsed = ts.createSourceFile(input, result, ts.ScriptTarget.Latest, true);
    assert.equal(parsed.parseDiagnostics.length,0);
    assert.equal(source.match(/function finishUpdateBaseSignatureDeclaration[^\r\n]+/)[0],
                 result.match(/function finishUpdateBaseSignatureDeclaration[^\r\n]+/)[0]);
    return result;
};
