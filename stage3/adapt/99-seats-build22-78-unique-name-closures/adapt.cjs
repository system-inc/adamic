"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");

const file = "src/compiler/factory/emitHelpers.ts";
const helperString = [
    "function helperString(input: TemplateStringsArray, ...args: string[]) {",
    "    return (uniqueName: EmitHelperUniqueNameCallback) => {",
    "        let result = \"\";",
    "        for (let i = 0; i < args.length; i++) {",
    "            result += input[i]!;",
    "            result += uniqueName(args[i]!);",
    "        }",
    "        result += input[input.length - 1]!;",
    "        return result;",
    "    };",
    "}",
].join("\r\n");
const snapshots = {
    asyncSuperHelper: [
        "",
        '            const ${"_superIndex"} = name => super[name];',
    ],
    advancedAsyncSuperHelper: [
        "",
        '            const ${"_superIndex"} = (function (geti, seti) {',
        "                const cache = Object.create(null);",
        "                return name => cache[name] || (cache[name] = { get value() { return geti(name); }, set value(v) { seti(name, v); } });",
        "            })(name => super[name], (name, value) => super[name] = value);",
    ],
};

function one(matches, description) {
    assert.equal(matches.length, 1, description);
    return matches[0];
}

function inspect(source) {
    assert.equal(ts.version, "6.0.3");
    const parsed = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true);
    assert.equal(parsed.parseDiagnostics.length, 0, file);
    const tag = one(parsed.statements.filter(node => ts.isFunctionDeclaration(node)
        && node.name?.text === "helperString"), "helperString declaration");
    assert.equal(tag.getText(parsed), helperString, "helperString: tag drift");
    const declarations = parsed.statements.filter(ts.isVariableStatement)
        .flatMap(statement => [...statement.declarationList.declarations]);
    const sites = Object.entries(snapshots).map(([name, lines]) => {
        const declaration = one(declarations.filter(node => ts.isIdentifier(node.name)
            && node.name.text === name), name);
        assert.equal(declaration.type?.getText(parsed), "EmitHelper", name);
        assert(declaration.initializer && ts.isObjectLiteralExpression(declaration.initializer), name);
        const property = one(declaration.initializer.properties.filter(node => ts.isPropertyAssignment(node)
            && ts.isIdentifier(node.name) && node.name.text === "text"), name + ".text");

        const value = property.initializer;
        const original = ts.isTaggedTemplateExpression(value);
        const template = original ? value.template : ts.isArrowFunction(value) ? value.body : undefined;
        assert(template && ts.isTemplateExpression(template), name + ": expected a template");
        assert.equal(template.templateSpans.length, 1, name + ": expected exactly one substitution");

        const before = "helperString" + "`" + lines.join("\r\n") + "`";
        const after = "(uniqueName) => " + before.slice("helperString".length)
            .replace('"_superIndex"', 'uniqueName("_superIndex")');
        assert.equal(value.getText(parsed), original ? before : after, name + ": snapshot drift");
        return {
            name, declaration: declaration.getText(parsed), original, before, after,
            start: value.getStart(parsed), end: value.end,
            cooked: [template.head.text, template.templateSpans[0].literal.text],
        };
    });
    assert(sites.every(site => site.original === sites[0].original), "partly adapted emitHelpers.ts");
    return { tag: tag.getText(parsed), sites };
}

function adapt(tree) {
    const filename = path.join(tree, file);
    const source = fs.readFileSync(filename, "utf8");
    const { sites } = inspect(source);
    if (!sites[0].original) return;

    let result = source;
    for (const site of [...sites].sort((left, right) => right.start - left.start)) {
        result = result.slice(0, site.start) + site.after + result.slice(site.end);
    }
    const adapted = inspect(result).sites;
    assert.deepEqual(adapted.map(site => site.cooked), sites.map(site => site.cooked));
    fs.writeFileSync(filename, result);
}

module.exports = { file, inspect };
if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
    adapt(path.resolve(process.argv[2]));
}
