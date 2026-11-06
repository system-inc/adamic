"use strict";
// Every textual `require(` in the files sys.ts and tsc.ts reach, with what the parser says each one is.
// usage: node require-census.cjs <tree> <output.json>
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const tree = path.resolve(process.argv[2]);
const roots = ["src/compiler/sys.ts", "src/tsc/tsc.ts"].map(name => path.join(tree, name));
const program = ts.createProgram(roots, { noEmit: true, target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler, allowImportingTsExtensions: true, types: [] });
// The checker binds every file, which sets the parent pointers the classification walks.
program.getTypeChecker();
const reachedFrom = {};
for (const root of roots) {
    const single = ts.createProgram([root], program.getCompilerOptions());
    for (const file of single.getSourceFiles()) if (file.fileName.startsWith(path.join(tree, "src/"))) (reachedFrom[file.fileName] ??= []).push(path.relative(tree, root));
}
const rows = [];
for (const file of program.getSourceFiles()) {
    if (!file.fileName.startsWith(path.join(tree, "src/"))) continue;
    const text = file.text;
    let at = text.indexOf("require(");
    while (at !== -1) {
        // Descend through every child, tokens and JSDoc included, to what covers the match: a child whose
        // leading trivia holds it (a comment), or the token whose own text holds it.
        let inner = file;
        let kind;
        for (;;) {
            const child = inner.getChildren(file).find(node => node.pos <= at && at < node.end);
            if (!child) break;
            if (at < child.getStart(file, true)) { kind = "comment"; break; }
            inner = child;
        }
        if (!kind) {
            if (inner.kind >= ts.SyntaxKind.FirstJSDocNode && inner.kind <= ts.SyntaxKind.LastJSDocNode || ts.findAncestor(inner, node => node.kind === ts.SyntaxKind.JSDoc)) kind = "comment";
            else if (ts.isStringLiteralLike(inner) || ts.isTemplateLiteralKind(inner.kind) || ts.isRegularExpressionLiteral(inner)) kind = "string";
            else if (ts.isIdentifier(inner) && inner.text === "require" && ts.isCallExpression(inner.parent) && inner.parent.expression === inner) {
                let enclosing = inner.parent;
                const chain = [];
                while (enclosing) { if (ts.isFunctionLike(enclosing) && enclosing.name) chain.push(enclosing.name.getText(file)); enclosing = enclosing.parent; }
                kind = `call require(${inner.parent.arguments.map(argument => argument.getText(file)).join(", ")}) in ${chain.join(" < ") || "module scope"}`;
                if (ts.findAncestor(inner, ts.isTypeNode)) kind += " (type)";
            }
            else kind = `token ${ts.SyntaxKind[inner.kind]} ${inner.getText(file).slice(0, 40)}`;
        }
        const position = file.getLineAndCharacterOfPosition(at);
        const lineStart = text.lastIndexOf("\n", at) + 1;
        const lineEnd = text.indexOf("\n", at);
        rows.push({ file: path.relative(tree, file.fileName), line: position.line + 1, kind, reachedFrom: reachedFrom[file.fileName] ?? [], text: text.slice(lineStart, lineEnd === -1 ? undefined : lineEnd).trim().slice(0, 160) });
        at = text.indexOf("require(", at + 1);
    }
}
const files = program.getSourceFiles().filter(file => file.fileName.startsWith(path.join(tree, "src/"))).map(file => path.relative(tree, file.fileName));
fs.writeFileSync(process.argv[3], JSON.stringify({ files: files.length, directories: [...new Set(files.map(name => path.dirname(name).split("/").slice(0, 2).join("/")))], rows }, null, 2) + "\n");
for (const row of rows) console.log(`${row.file}:${row.line}\t${row.kind}\t${row.text}`);
console.log(`${rows.length} hits in ${files.length} files`);
