"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const file = "src/compiler/builder.ts";
const owner = "handleDtsMayChangeOfFileAndExportsOfFile";
const receiver = "state.referencedMap!.getKeys(filePath)";
const workerCall = [
    owner + "(",
    "            state,",
    "            referencingFilePath,",
    "            invalidateJsFiles,",
    "            seenFileAndExportsOfFile,",
    "            cancellationToken,",
    "            host,",
    "        )",
].join("\r\n");
const before = `${receiver}?.forEach(referencingFilePath =>\r\n        ${workerCall}\r\n    );`;
const afterWorkerCall = workerCall;
const after = `${receiver}?.forEach(referencingFilePath =>\r\n        { ${afterWorkerCall}; }\r\n    );`;

function locate(source) {
    assert.equal(source.parseDiagnostics.length, 0, "invalid builder source");
    const owners = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === owner);
    assert.equal(owners.length, 1, "missing or duplicate worker declaration");
    const statements = owners[0].body?.statements.filter(node =>
        ts.isExpressionStatement(node) && ts.isCallExpression(node.expression) &&
        ts.isPropertyAccessExpression(node.expression.expression) &&
        node.expression.expression.name.text === "forEach");
    assert.equal(statements?.length, 1, "missing or duplicate Set visit statement");
    const statement = statements[0];
    const current = statement.getText(source);
    assert(current === before || current === after, "Set visit statement drift");
    const arrow = statement.expression.arguments[0];
    assert(ts.isArrowFunction(arrow));
    return { statement, arrow, current };
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const filePath = path.join(tree, file);
    const text = fs.readFileSync(filePath, "utf8");
    const program = ts.createProgram([filePath], {
        target: ts.ScriptTarget.ES2020,
        module: ts.ModuleKind.NodeNext,
        strict: true,
        skipLibCheck: true,
        types: [],
        noEmit: true,
    });
    const source = program.getSourceFile(filePath);
    const { statement, arrow, current } = locate(source);
    const checker = program.getTypeChecker();
    const method = statement.expression.expression;
    assert.equal(checker.typeToString(checker.getTypeAtLocation(method.expression)),
        "ReadonlySet<Path> | undefined", "Set receiver type drift");
    const signature = checker.getResolvedSignature(statement.expression);
    assert.match(signature?.declaration?.getSourceFile().fileName ?? "",
        /\/lib\.es2015\.collection\.d\.ts$/, "forEach does not resolve to the stock Set method");
    assert.equal(checker.typeToString(checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(arrow))),
        current === before ? "boolean | undefined" : "void", "Set callback return drift");
    if (current === after) return { files: 0, edits: 0 };

    const revised = text.slice(0, statement.getStart(source)) + after + text.slice(statement.end);
    locate(ts.createSourceFile(filePath, revised, ts.ScriptTarget.Latest, true));
    assert.equal(fs.readFileSync(filePath, "utf8"), text, "source changed during planning");
    fs.writeFileSync(filePath, revised);
    return { files: 1, edits: 1 };
}

module.exports = { adapt, file, locate, before, after };
if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs TREE");
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
