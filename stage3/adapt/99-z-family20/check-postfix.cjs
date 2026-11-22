const assert = require("node:assert/strict");

function checkPostfix(factory, first = {}, second = {}) {
    for (const name of ["JSDocNonNullableType", "JSDocNullableType"]) {
        const create = factory["create" + name];
        const update = factory["update" + name];
        assert.equal(factory["create" + name], create, name + " create getter identity");
        assert.equal(factory["update" + name], update, name + " update getter identity");
        for (const postfix of [true, false, undefined]) {
            const node = create(first, postfix);
            assert.equal(node.postfix, postfix ?? false, name + " postfix is assigned on creation");
            assert.equal(update(node, node.type), node, name + " no-op update identity");
            const changed = update(node, second);
            assert.notEqual(changed, node, name + " changed update identity");
            assert.equal(changed.type, second, name + " changed child");
            assert.equal(changed.postfix, postfix ?? false, name + " postfix is preserved on update");
        }
    }
}

function checkSource(ts, filename, source, coreSource) {
    const parse = (name, text) => ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true);
    const file = parse(filename, source);
    const core = parse("core.ts", coreSource);
    assert.equal(file.parseDiagnostics.length, 0, filename);
    const factory = file.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === "createNodeFactory");
    const members = factory.body.statements.find(node => ts.isVariableStatement(node) &&
        node.declarationList.declarations.some(decl => decl.name.getText(file) === "factory"))
        .declarationList.declarations.find(decl => decl.name.getText(file) === "factory").initializer;
    const getters = members.properties.filter(node => /^(?:create|update)JSDoc(?:NonNullable|Nullable)Type$/.test(node.name?.getText(file)));
    assert.equal(getters.length, 4, "nullable and non-nullable factory getters");
    const declarations = factory.body.statements.filter(node => {
        if (ts.isFunctionDeclaration(node)) {
            return /^(?:createBaseNode|update|(?:create|update)JSDoc(?:PrePostfix)?UnaryTypeWorkerForJSDoc(?:NonNullable|Nullable)Type)$/.test(node.name?.text);
        }
        return ts.isVariableStatement(node) && node.declarationList.declarations.some(decl =>
            /^getJSDocPrePostfixUnaryType(?:Create|Update)FunctionForJSDoc(?:NonNullable|Nullable)Type$/.test(decl.name.getText(file)));
    });
    assert.equal(declarations.length, 12, "specialized postfix workers, memoizers and dependencies");
    const memoize = core.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === "memoizeOne");
    const snippet = [
        memoize.getText(core).replace(/^export /, ""),
        ...declarations.map(node => node.getText(file)),
        "return ({" + getters.map(node => node.getText(file)).join(",\n") + "});",
    ].join("\n");
    const js = ts.transpileModule(snippet, {compilerOptions: {target: ts.ScriptTarget.ES2020}}).outputText;
    // Execute the adapted getters and workers with small allocator and metadata fixtures.
    const makeFactory = new Function("SyntaxKind", "baseFactory", "parenthesizerRules", "setOriginal", "setTextRange", js);
    const identity = value => value;
    checkPostfix(makeFactory(ts.SyntaxKind, {createBaseNode: kind => ({kind})},
        () => ({parenthesizeNonArrayTypeOfPostfixType: identity}), identity, identity));
}

module.exports = {checkPostfix, checkSource};

if (require.main === module) {
    const [modulePath, parserPath] = process.argv.slice(2);
    const {factory} = require(modulePath);
    const {SyntaxKind} = require(parserPath);
    checkPostfix(factory,
        factory.createKeywordTypeNode(SyntaxKind.StringKeyword),
        factory.createKeywordTypeNode(SyntaxKind.NumberKeyword));
    console.log("Nullable and non-nullable factory postfix, memoization and update identity verified");
}
