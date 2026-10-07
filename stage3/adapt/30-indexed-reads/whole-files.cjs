"use strict";
const entries = require("./whole-sites.json");
function planWhole(ts, file, text, check = false) {
    if (!entries.some(entry => entry.file === file && !entry.action.startsWith("decline"))) return { text, edits: 0 };
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (file === "core.ts") {
        const before = "for (const elements of arrayFrom(multiMap.values()))";
        const after = "for (const elements of arrayFrom<TElement | TElement[]>(multiMap.values()))";
        const newline = text.includes("\r\n") ? "\r\n" : "\n";
        const marker = newline + "/** @internal */" + newline + "export function isNodeLikeSystem(): boolean {";
        const host = "declare const process: { nextTick?: unknown; browser?: unknown; } | undefined; declare const require: unknown;" + marker;
        let edits = 0;
        for (const [original, adapted] of [[before, after], [marker, host]]) {
            if (text.split(adapted).length === 2) continue;
            if (text.split(original).length !== 2) throw new Error("core closure owner drift: " + original);
            if (check) throw new Error("truthful core declaration missing: " + original);
            text = text.replace(original, adapted); edits++;
        }
        return { text, edits };
    }
    if (file === "utilities.ts") {
        const repairs = [
            ["return arrayFrom(directivesByLine.entries())", "return arrayFrom<[string, CommentDirective]>(directivesByLine.entries())"],
            ["let nextCode: number = codes[i];", "let nextCode: number | undefined = codes[i];"],
            ["while ((nextCode & 0B11000000) === 0B10000000)", "while (nextCode !== undefined && (nextCode & 0B11000000) === 0B10000000)"],
            ["const stringReplace = String.prototype.replace;", "const stringReplace: (this: string, searchValue: string, replaceValue: string) => string = String.prototype.replace;"],
        ];
        let edits = 0;
        for (const [before, after] of repairs) {
            const original = text.split(before).length - 1, adapted = text.split(after).length - 1;
            if (adapted === 1 && original === 0) continue;
            if (original !== 1 || adapted !== 0) throw new Error("utilities closure owner drift: " + before);
            if (check) throw new Error("truthful utilities declaration or narrowing missing: " + before);
            text = text.replace(before, after); edits++;
        }
        return { text, edits };
    }
    if (file === "parser.ts") {
        const repairs = [
            ["currentNode(position: number): Node;", "currentNode(position: number): Node | undefined;"],
            ['function getNamedPragmaArguments(pragma: PragmaDefinition, text: string | undefined): { [index: string]: string; } | "fail"', 'function getNamedPragmaArguments(pragma: PragmaDefinition, text: string | undefined): { [index: string]: string | undefined; } | "fail"'],
            ["const argMap: { [index: string]: string; } = {};", "const argMap: { [index: string]: string | undefined; } = {};"],
            ["if (nodeIsMissing(node) || intersectsIncrementalChange(node) || containsParseError(node)) {", "const missingNode = nodeIsMissing(node); if (missingNode || node === undefined || intersectsIncrementalChange(node) || containsParseError(node)) {"],
        ];
        let edits = 0;
        for (const [before, after] of repairs) {
            const original = text.split(before).length - 1, adapted = text.split(after).length - 1;
            if (adapted === 1 && original === 0) continue;
            if (original !== 1 || adapted !== 0) throw new Error("parser closure owner drift: " + before);
            if (check) throw new Error("truthful parser declaration or narrowing missing: " + before);
            text = text.replace(before, after); edits++;
        }
        return { text, edits };
    }
    if (file === "scanner.ts") {
        const fn = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === "computePositionOfLineAndCharacter");
        if (!fn || fn.parameters.map(node => node.name.getText(source)).join(",") !== "lineStarts,line,character,debugText,allowEdits") throw new Error("line clamp owner drift");
        if (fn.parameters[0].type.getText(source) !== "readonly number[]" || fn.parameters[2].type.getText(source) !== "number" ||
            !fn.body.statements.some(node => ts.isVariableStatement(node) && node.getText(source) === "const res = lineStarts[line]! + character;")) throw new Error("line clamp primitive res proof drift");
        const branch = fn.body.statements.find(node => ts.isIfStatement(node) && node.expression.getText(source) === "allowEdits");
        if (!branch || !ts.isBlock(branch.thenStatement)) throw new Error("line clamp branch drift");
        const statements = branch.thenStatement.statements;
        const tail = 'typeof debugText === "string" && res > debugText.length ? debugText.length : res;';
        const original = "return res > lineStarts[line + 1] ? lineStarts[line + 1]! : " + tail;
        const adapted = "return nextLineStart !== undefined && res > nextLineStart ? lineStarts[line + 1]! : " + tail;
        if (statements.length === 2 && statements[0].getText(source) === "const nextLineStart = lineStarts[line + 1];" && statements[1].getText(source) === adapted) return { text, edits: 0 };
        if (statements.length !== 1 || !ts.isReturnStatement(statements[0]) || statements[0].getText(source) !== original) throw new Error("line clamp read structure drift");
        if (check) throw new Error("explicit next-line narrowing missing: scanner.ts:491");
        const at = statements[0].getStart(source), end = statements[0].end;
        const newline = text.includes("\r\n") ? "\r\n" : "\n";
        return { text: text.slice(0, at) + "const nextLineStart = lineStarts[line + 1]; " + adapted + text.slice(end), edits: 1 };
    }
    const found = [];
    function visit(node) {
        if (ts.isVariableDeclaration(node) && node.name.getText(source) === "assertionCache") found.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    if (file !== "debug.ts" || found.length !== 1) throw new Error("unknown whole-file closure: " + file);
    const variable = found[0], partial = variable.type;
    if (!ts.isTypeReferenceNode(partial) || partial.typeName.getText(source) !== "Partial" || partial.typeArguments.length !== 1) throw new Error("assertion cache Partial drift");
    const record = partial.typeArguments[0];
    if (!ts.isTypeReferenceNode(record) || record.typeName.getText(source) !== "Record" || record.typeArguments.length !== 2 || record.typeArguments[0].getText(source) !== "AssertionKeys") throw new Error("assertion cache Record drift");
    const value = record.typeArguments[1];
    const adapted = ts.isUnionTypeNode(value) && value.types.length === 2 && value.types[1].kind === ts.SyntaxKind.UndefinedKeyword;
    const object = adapted ? value.types[0] : value;
    if (!ts.isTypeLiteralNode(object) || object.members.length !== 2 || object.members.map(member => member.getText(source)).join(" ") !== "level: AssertionLevel; assertion: AnyFunction;") throw new Error("assertion cache entry drift");
    if (adapted) return { text, edits: 0 };
    if (check) throw new Error("truthful cache value union missing: debug.ts:155");
    return { text: text.slice(0, object.end) + " | undefined" + text.slice(object.end), edits: 1 };
}
module.exports = { planWhole };
