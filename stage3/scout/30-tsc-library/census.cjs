// Count compiler calls, keeping source sites and emitted helper text separate.
// CENSUS_TYPESCRIPT names the stock TypeScript package used for the AST and types.
const fs = require("node:fs");
const path = require("node:path");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const root = path.resolve(process.argv[2]);
const files = fs
  .readdirSync(root, { recursive: true })
  .filter(
    (x) =>
      x.endsWith(".ts") && !x.endsWith(".d.ts") && !x.includes(".generated."),
  )
  .map((x) => path.join(root, x))
  .sort();
const config = ts.readConfigFile(
  path.join(root, "tsconfig.json"),
  ts.sys.readFile,
);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, root);
const program = ts.createProgram(parsed.fileNames, {
  ...parsed.options,
  noEmit: true,
});
const checker = program.getTypeChecker();
const sites = [],
  helpers = [],
  instances = [],
  borrowed = [];
const families = new Set([
  "Number",
  "Math",
  "Object",
  "JSON",
  "Map",
  "Set",
  "String",
  "Boolean",
]);
const globals = new Set([
  "parseInt",
  "parseFloat",
  "isNaN",
  "isFinite",
  "String",
  "Number",
  "Boolean",
  "Object",
]);
function unwrap(node) {
  while (
    ts.isParenthesizedExpression(node) ||
    ts.isAsExpression(node) ||
    ts.isTypeAssertionExpression(node) ||
    ts.isNonNullExpression(node)
  )
    node = node.expression;
  return node;
}
function scan(file, list, offset = 0, types = true) {
  function visit(node) {
    if (
      types &&
      (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) &&
      ts.isPropertyAssignment(node.parent) &&
      node.parent.name.getText(file) === "text"
    ) {
      const helper = ts.createSourceFile(
        "helper.js",
        node.text,
        ts.ScriptTarget.Latest,
        true,
        ts.ScriptKind.JS,
      );
      scan(
        helper,
        helpers,
        file.getLineAndCharacterOfPosition(node.getStart(file)).line,
        false,
      );
    }
    if (ts.isCallExpression(node)) {
      let callee = unwrap(node.expression),
        name,
        instance = false,
        cached = false;
      if (ts.isIdentifier(callee) && globals.has(callee.text))
        name = callee.text;
      if (
        ts.isPropertyAccessExpression(callee) ||
        ts.isElementAccessExpression(callee)
      ) {
        const member = ts.isPropertyAccessExpression(callee)
          ? callee.name.text
          : ts.isStringLiteral(callee.argumentExpression)
            ? callee.argumentExpression.text
            : null;
        const receiver = unwrap(callee.expression);
        if (ts.isIdentifier(receiver) && families.has(receiver.text))
          name = receiver.text + "." + member;
        else if (
          types &&
          member === "call" &&
          ts.isIdentifier(receiver) &&
          receiver.text === "hasOwnProperty"
        ) {
          const declaration =
            checker.getSymbolAtLocation(receiver)?.valueDeclaration;
          if (
            declaration &&
            ts.isVariableDeclaration(declaration) &&
            declaration.initializer?.getText() ===
              "Object.prototype.hasOwnProperty"
          ) {
            name = "Object.prototype.hasOwnProperty.call";
            cached = true;
          }
        } else if (
          member === "apply" &&
          receiver.getText(file) === "String.fromCharCode"
        )
          name = "String.fromCharCode.apply";
        else if (types && member) {
          const receiver = checker.getTypeAtLocation(callee.expression);
          const parts = receiver.isUnion() ? receiver.types : [receiver];
          if (
            parts.some((x) =>
              ["Map", "ReadonlyMap", "Set", "ReadonlySet"].includes(
                x.symbol?.name,
              ),
            )
          ) {
            name = "Map/Set instance." + member;
            instance = true;
          }
        }
      }
      if (name) {
        const pos = file.getLineAndCharacterOfPosition(node.getStart(file));
        const args = node.arguments.map((x) =>
          types
            ? checker.typeToString(checker.getTypeAtLocation(x))
            : ts.SyntaxKind[x.kind],
        );
        const shape = name + "(" + args.join(", ") + ")";
        const site = {
          file: types
            ? path.relative(root, file.fileName)
            : "factory/emitHelpers.ts",
          line: pos.line + 1 + offset,
          column: pos.character + 1,
          name,
          shape,
          arguments: args,
          spread: node.arguments.some(ts.isSpreadElement),
          text: node.getText(file).replace(/\s+/g, " "),
        };
        (cached ? borrowed : instance ? instances : list).push(site);
      }
    }
    ts.forEachChild(node, visit);
  }
  visit(file);
}
for (const file of program.getSourceFiles())
  if (files.includes(file.fileName)) scan(file, sites);
function counts(values, key) {
  const result = {};
  for (const value of values)
    result[value[key]] = (result[value[key]] || 0) + 1;
  return result;
}
const result = {
  files: files.length,
  typescript: ts.version,
  counts: counts(sites, "name"),
  shapes: counts(sites, "shape"),
  instanceCounts: counts(instances, "name"),
  helperCounts: counts(helpers, "name"),
  borrowedCounts: counts(borrowed, "name"),
  sites,
  instances,
  helpers,
  borrowed,
};
const output = JSON.stringify(result, null, 2) + "\n";
if (process.argv[3]) fs.writeFileSync(process.argv[3], output);
else process.stdout.write(output);
