#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");
const { prove } = require("./proof.cjs");

function adapt(tree, ts) {
    if (ts.version !== "6.0.3") throw new Error(`want stock typescript@6.0.3, got ${ts.version}`);
    const directory = path.join(tree, "src/compiler");
    const files = ["checker", "debug", "emitter", "parser", "semver", "sourcemap", "utilities", "utilitiesPublic", "types"];
    const sources = new Map(files.map(file => {
        const name = path.join(directory, file + ".ts");
        const source = ts.createSourceFile(name, fs.readFileSync(name, "utf8"), ts.ScriptTarget.Latest, true);
        if (source.parseDiagnostics.length) throw new Error(`cannot parse ${name}`);
        return [file, source];
    }));
    const edits = new Map();
    const evidence = [];
    const declined = [];
    function nodes(root, predicate) {
        const found = [];
        function visit(node) {
            if (predicate(node)) found.push(node);
            ts.forEachChild(node, visit);
        }
        visit(root);
        return found;
    }
    function one(root, predicate, description) {
        const found = nodes(root, predicate);
        if (found.length !== 1) throw new Error(`${description}: want one node, got ${found.length}`);
        return found[0];
    }
    function func(file, name) {
        return one(sources.get(file), node => ts.isFunctionDeclaration(node) && node.name?.text === name, `${file}:${name}`);
    }
    function variable(root, name) {
        return one(root, node => ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.name.text === name, name);
    }
    function unwrapped(node) {
        while (ts.isParenthesizedExpression(node) || ts.isNonNullExpression(node)) node = node.expression;
        return node;
    }
    function access(node, name, index) {
        node = unwrapped(node);
        return ts.isElementAccessExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === name &&
            ts.isNumericLiteral(node.argumentExpression) && Number(node.argumentExpression.text) === index;
    }
    function edit(source, node, replacement, reason) {
        const start = node.getStart(source);
        const original = node.getText(source);
        if (original === replacement) return;
        let list = edits.get(source);
        if (!list) edits.set(source, list = []);
        list.push({ start, end: node.end, replacement });
        const where = source.getLineAndCharacterOfPosition(start);
        evidence.push({ file: path.basename(source.fileName), line: where.line + 1, original, replacement, reason });
    }
    function assertPresent(source, expression, reason) {
        if (ts.isNonNullExpression(expression.parent)) return;
        const text = expression.getText(source);
        edit(source, expression, ts.isBinaryExpression(expression) ? `(${text})!` : `${text}!`, reason);
    }
    function literalProof(file, name) {
        const source = sources.get(file);
        const declaration = variable(source, name);
        const literal = declaration.initializer;
        if (!literal || !ts.isRegularExpressionLiteral(literal)) throw new Error(`want regex literal ${name}`);
        return regexProof(literal.getText(source));
    }
    function regexProof(text) {
        const last = text.lastIndexOf("/");
        if (text[0] !== "/" || last < 1) throw new Error("want regex literal");
        if (!/^[gim]*$/.test(text.slice(last + 1))) throw new Error("declined unsupported regex flags");
        const result = prove(text.slice(1, last));
        if (!result.supported) throw new Error(`declined regex: ${text}: ${result.reason}`);
        return { ...result, regex: text };
    }
    function required(proof, group, description) {
        if (!proof.mandatory.includes(group)) {
            declined.push({ site: description, regex: proof.regex, group, reason: "capture can be bypassed" });
            return false;
        }
        return true;
    }
    function capture(file, root, receiver, group, proof, description) {
        const source = sources.get(file);
        const node = one(root, node => ts.isElementAccessExpression(node) && access(node, receiver, group), description);
        if (required(proof, group, description)) assertPresent(source, node, `mandatory group ${group} of ${proof.regex}`);
    }
    // Match identities are taken from their initializer in the parsed current tree.
    const debug = func("debug", "getFunctionName");
    const debugMatch = variable(debug, "match").initializer;
    if (!ts.isCallExpression(debugMatch) || !ts.isPropertyAccessExpression(debugMatch.expression) ||
        debugMatch.expression.name.text !== "exec" || !ts.isRegularExpressionLiteral(debugMatch.expression.expression)) {
        throw new Error("unexpected getFunctionName match initializer");
    }
    capture("debug", debug, "match", 1, regexProof(debugMatch.expression.expression.getText(sources.get("debug"))), "getFunctionName name");

    const extract = func("parser", "extractPragmas");
    capture("parser", extract, "tripleSlash", 1, literalProof("parser", "tripleSlashXMLCommentStartRegEx"), "XML pragma name");
    const constructor = variable(func("parser", "getNamedArgRegEx"), "result").initializer;
    const template = constructor?.arguments?.[0];
    if (!ts.isNewExpression(constructor) || constructor.expression.getText() !== "RegExp" ||
        !template || !ts.isTemplateExpression(template) || template.templateSpans.length !== 1 ||
        template.templateSpans[0].expression.getText() !== "name" || constructor.arguments[1]?.text !== "im" ||
        template.head.text !== "(\\s" || template.templateSpans[0].literal.text !== "\\s*=\\s*)(?:(?:'([^']*)')|(?:\"([^\"]*)\"))") {
        throw new Error("unexpected named-argument constructor");
    }
    const pragmas = variable(sources.get("types"), "commentPragmas");
    const names = nodes(pragmas, node => ts.isPropertyAssignment(node) && node.name.getText() === "name");
    if (!names.length || names.some(node => !ts.isStringLiteral(node.initializer) || !/^[a-z-]+$/.test(node.initializer.text))) {
        throw new Error("pragma argument names may inject regex syntax");
    }
    // Density does not prove index bounds. The positional helper handles
    // required missing arguments before its assignment; verify that its actual
    // non-XML callers have no optional arguments that bypass that guard.
    const specifications = ts.isAsExpression(pragmas.initializer) ? pragmas.initializer.expression : pragmas.initializer;
    if (!ts.isObjectLiteralExpression(specifications)) throw new Error("unreviewed pragma specifications");
    for (const specification of specifications.properties) {
        if (!ts.isPropertyAssignment(specification) || !ts.isObjectLiteralExpression(specification.initializer)) throw new Error("unreviewed pragma definition");
        const properties = specification.initializer.properties;
        if (properties.some(node => ts.isPropertyAssignment(node) && node.name.getText() === "kind" &&
            node.initializer.getText() === "PragmaKindFlags.TripleSlashXML")) continue;
        const argumentsProperty = properties.find(node => ts.isPropertyAssignment(node) && node.name.getText() === "args");
        if (!argumentsProperty) continue;
        if (!ts.isArrayLiteralExpression(argumentsProperty.initializer)) throw new Error("unreviewed positional pragma arguments");
        for (const argument of argumentsProperty.initializer.elements) {
            if (!ts.isObjectLiteralExpression(argument) || argument.properties.some(node => ts.isPropertyAssignment(node) &&
                node.name.getText() === "optional" && node.initializer.kind !== ts.SyntaxKind.FalseKeyword)) {
                throw new Error("declined optional positional pragma argument");
            }
        }
    }
    const namedProof = { ...prove(template.head.text + "name" + template.templateSpans[0].literal.text), regex: "getNamedArgRegEx fixed template" };
    if (!namedProof.supported) throw new Error("unsupported named-argument expression");
    capture("parser", extract, "matchResult", 1, namedProof, "XML argument prefix");
    // User-authorized latent-bug boundary. Neither alternative capture is asserted.
    const value = variable(extract, "value").initializer;
    const valueExpression = unwrapped(value);
    if (!ts.isBinaryExpression(valueExpression) || !([ts.SyntaxKind.BarBarToken, ts.SyntaxKind.QuestionQuestionToken].includes(valueExpression.operatorToken.kind)) ||
        !access(valueExpression.left, "matchResult", 2) || !access(valueExpression.right, "matchResult", 3)) {
        throw new Error("unexpected XML argument value expression");
    }
    if (valueExpression.operatorToken.kind === ts.SyntaxKind.QuestionQuestionToken && !ts.isNonNullExpression(value)) throw new Error("unreviewed nullish value without checked boundary");
    if (!ts.isNonNullExpression(value)) assertPresent(sources.get("parser"), valueExpression, "authorized checked boundary; empty single quotes remain a latent bug");
    const single = literalProof("parser", "singleLinePragmaRegEx");
    const multiline = one(extract, node => ts.isVariableDeclaration(node) && node.name.getText() === "multiLinePragmaRegEx", "multiline pragma regex");
    const multi = regexProof(multiline.initializer.getText(sources.get("parser")));
    if (required(single, 1, "single-line pragma name") && required(multi, 1, "multi-line pragma name")) {
        capture("parser", func("parser", "addPragmaForMatch"), "match", 1, single, "shared pragma name");
    }

    const sourceMap = literalProof("sourcemap", "sourceMapCommentRegExp");
    capture("sourcemap", func("sourcemap", "tryGetSourceMappingURL"), "comment", 1, sourceMap, "source map URL");
    const locale = func("utilitiesPublic", "validateLocaleAndSetLanguage");
    const localeMatch = variable(locale, "matchResult").initializer;
    if (!ts.isCallExpression(localeMatch) || !ts.isPropertyAccessExpression(localeMatch.expression) ||
        localeMatch.expression.name.text !== "exec" || !ts.isRegularExpressionLiteral(localeMatch.expression.expression)) {
        throw new Error("unexpected locale match initializer");
    }
    capture("utilitiesPublic", locale, "matchResult", 1, regexProof(localeMatch.expression.expression.getText(sources.get("utilitiesPublic"))), "locale language");

    // Bind required major once. All optional captures retain their exact defaults.
    for (const [functionName, regexName] of [["tryParseComponents", "versionRegExp"], ["parsePartial", "partialRegExp"]]) {
        const source = sources.get("semver");
        const root = func("semver", functionName);
        const proof = literalProof("semver", regexName);
        const binding = one(root, node => ts.isVariableDeclaration(node) && ts.isArrayBindingPattern(node.name), `${functionName} binding`);
        const major = binding.name.elements[1];
        const majorDeclaration = nodes(root, node => ts.isVariableDeclaration(node) && node.name.getText() === "major");
        if (!required(proof, 1, `${functionName} major`)) continue;
        if (ts.isOmittedExpression(major) && majorDeclaration.length === 1) {
            if (!ts.isNonNullExpression(majorDeclaration[0].initializer) || !access(majorDeclaration[0].initializer, "match", 1)) throw new Error("unexpected adapted major");
            continue;
        }
        if (!ts.isBindingElement(major) || major.name.getText() !== "major" || major.initializer || majorDeclaration.length ||
            binding.initializer.getText() !== "match") throw new Error("unexpected major binding");
        const statement = binding.parent.parent;
        if (!ts.isVariableStatement(statement) || statement.declarationList.declarations.length !== 1) throw new Error("unexpected binding statement");
        const original = statement.getText(source);
        const offset = major.getStart(source) - statement.getStart(source);
        const withoutMajor = original.slice(0, offset) + original.slice(offset + major.getWidth(source));
        const indentation = source.text.slice(source.text.lastIndexOf("\n", statement.getStart(source)) + 1, statement.getStart(source));
        const newline = source.text.includes("\r\n") ? "\r\n" : "\n";
        edit(source, statement, `const major = match[1]!;${newline}${indentation}${withoutMajor}`, `mandatory major of ${proof.regex}; defaults unchanged`);
    }
    const ranges = func("semver", "parseRange");
    const hyphen = literalProof("semver", "hyphenRegExp");
    const range = literalProof("semver", "rangeRegExp");
    const calls = nodes(ranges, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression) && ["parseHyphen", "parseComparator"].includes(node.expression.text));
    if (calls.length !== 2) throw new Error("unexpected range parse calls");
    for (const call of calls) {
        const groups = call.expression.text === "parseHyphen" ? [1, 2] : [2];
        const proof = call.expression.text === "parseHyphen" ? hyphen : range;
        for (const group of groups) {
            const argument = call.arguments[group - 1];
            if (!access(argument, "match", group)) throw new Error("unexpected range argument");
            if (required(proof, group, `${call.expression.text} group ${group}`) && !ts.isNonNullExpression(argument)) {
                assertPresent(sources.get("semver"), argument, `mandatory group ${group} of ${proof.regex}`);
            }
        }
    }
    const comparator = func("semver", "parseComparator");
    if (!nodes(comparator, node => ts.isCaseClause(node) && node.expression.getText() === "undefined").length) throw new Error("missing absent-operator handler");
    const operator = comparator.parameters[0];
    if (operator.name.getText() !== "operator" || !["string", "string | undefined"].includes(operator.type.getText())) throw new Error("unexpected operator parameter");
    edit(sources.get("semver"), operator.type, "string | undefined", "optional range operator already handled by case undefined");

    // Only the six reviewed capture-free split expressions receive a density assertion.
    const splits = [
        ["checker", "line.replace(/^\\s+/, \" \")", "callback"],
        ["utilities", "line.replace(/^\\s*\\*/, \"\").trimStart()", "callback"],
        ["emitter", "emitJSDoc", "array"],
        ["emitter", "writeLines", "array"],
        ["parser", "getNamedPragmaArguments", "array"],
        ["semver", "parseRange", "array"],
    ];
    for (const [file, scope, mode] of splits) {
        const source = sources.get(file);
        if (mode === "callback") {
            const arrow = one(source, node => ts.isArrowFunction(node) && node.body.getText().replace(/line!/g, "line") === scope, `${file} line callback`);
            const map = arrow.parent;
            const split = map.expression?.expression;
            if (!ts.isCallExpression(map) || !ts.isPropertyAccessExpression(map.expression) || map.expression.name.text !== "map" ||
                !split || !ts.isCallExpression(split) || !ts.isPropertyAccessExpression(split.expression) || split.expression.name.text !== "split" ||
                !ts.isRegularExpressionLiteral(split.arguments[0])) throw new Error("unexpected line split callback");
            if (regexProof(split.arguments[0].getText(source)).captures !== 0) throw new Error("declined split with captures");
            const receiver = one(arrow.body, node => ts.isPropertyAccessExpression(node) && node.name.text === "replace", "line replace").expression;
            if (!ts.isIdentifier(unwrapped(receiver)) || unwrapped(receiver).text !== "line") throw new Error("unexpected callback receiver");
            if (!ts.isNonNullExpression(receiver)) assertPresent(source, receiver, "capture-free separator yields only strings");
        }
        else {
            const root = func(file, scope);
            const split = one(root, node => ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) && node.expression.name.text === "split" &&
                (ts.isRegularExpressionLiteral(node.arguments[0]) || node.arguments[0]?.getText() === "whitespaceRegExp"), `${file}:${scope} split`);
            const proof = ts.isRegularExpressionLiteral(split.arguments[0]) ? regexProof(split.arguments[0].getText(source)) : literalProof(file, "whitespaceRegExp");
            if (proof.captures !== 0) throw new Error(`declined split with captures: ${proof.regex}`);
            if (ts.isAsExpression(split.parent)) {
                if (split.parent.type.getText() !== "string[]") throw new Error("unexpected existing split assertion");
            }
            else edit(source, split, `${split.getText(source)} as string[]`, `capture-free separator ${proof.regex} yields a dense string array`);
        }
    }
    // Declines are all-or-nothing: a mutant must not leave a partly adapted tree.
    if (declined.length) throw new Error(JSON.stringify({ declined }));
    for (const [source, list] of edits) {
        list.sort((left, right) => right.start - left.start);
        let text = source.text;
        let boundary = text.length;
        for (const change of list) {
            if (change.end > boundary) throw new Error(`overlapping edits in ${source.fileName}`);
            text = text.slice(0, change.start) + change.replacement + text.slice(change.end);
            boundary = change.start;
        }
        if (fs.readFileSync(source.fileName, "utf8") !== source.text) throw new Error("source changed while planning");
        const parsed = ts.createSourceFile(source.fileName, text, ts.ScriptTarget.Latest, true);
        if (parsed.parseDiagnostics.length) throw new Error(`adapted syntax error in ${source.fileName}`);
        edits.set(source, text);
    }
    for (const [source, text] of edits) fs.writeFileSync(source.fileName, text);
    return { files: edits.size, edits: evidence.length, evidence, declined: [
        { site: "XML value groups 2 and 3", reason: "alternative branches; whole-expression checked boundary authorized instead" },
        { site: "range operator group 1", reason: "optional; truthful parameter includes undefined" },
    ] };
}

module.exports = { adapt };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    }
    catch (error) { console.error(error.message); process.exitCode = 1; }
}
