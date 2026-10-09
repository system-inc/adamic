// AST census of executable compiler operations and separately parsed emitted helper text.
// Run with CENSUS_TYPESCRIPT pointing at stock TypeScript 6.0.3; optional CENSUS_TYPE_ROOTS supplies pinned @types declarations.
const fs = require("node:fs"),
  path = require("node:path"),
  crypto = require("node:crypto"),
  ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
if (!process.argv[2])
  throw new Error(
    "usage: node census.cjs <TypeScript checkout or compiler directory> [output.json]",
  );
const source = path.resolve(process.argv[2]);
const root = fs.existsSync(path.join(source, "src/compiler"))
  ? path.join(source, "src/compiler")
  : fs.existsSync(path.join(source, "tsc/testdata/fixtures/compiler"))
    ? path.join(source, "tsc/testdata/fixtures/compiler")
    : source;
const output = process.argv[3];
function walk(dir) {
  return fs
    .readdirSync(dir, { withFileTypes: true })
    .flatMap((e) =>
      e.isDirectory()
        ? walk(path.join(dir, e.name))
        : e.name.endsWith(".ts") &&
            !e.name.endsWith(".d.ts") &&
            !e.name.includes(".generated.")
          ? [path.join(dir, e.name)]
          : [],
    );
}
const files = walk(root).sort();
const configName = path.join(root, "tsconfig.json");
const read = ts.readConfigFile(configName, ts.sys.readFile);
if (read.error)
  throw new Error(
    ts.flattenDiagnosticMessageText(read.error.messageText, "\n"),
  );
const parsed = ts.parseJsonConfigFileContent(read.config, ts.sys, root);
if (parsed.errors.length)
  throw new Error(
    parsed.errors
      .map((e) => ts.flattenDiagnosticMessageText(e.messageText, "\n"))
      .join("\n"),
  );
const options = {
  ...parsed.options,
  noEmit: true,
  ...(process.env.CENSUS_TYPE_ROOTS
    ? { typeRoots: process.env.CENSUS_TYPE_ROOTS.split(path.delimiter) }
    : {}),
};
const p = ts.createProgram(parsed.fileNames, options);
const c = p.getTypeChecker();
const sites = [];
const helpers = [];
function possibleObject(t) {
  if (t.isUnion()) return t.types.some(possibleObject);
  if (t.flags & ts.TypeFlags.TypeParameter) {
    const base = c.getBaseConstraintOfType(t);
    if (base && base !== t) return possibleObject(base);
  }
  return !!(
    t.flags &
    (ts.TypeFlags.Object |
      ts.TypeFlags.NonPrimitive |
      ts.TypeFlags.TypeParameter |
      ts.TypeFlags.Any |
      ts.TypeFlags.Unknown)
  );
}
function definiteObject(t) {
  if (t.isUnion()) return t.types.some(definiteObject);
  return !!(t.flags & (ts.TypeFlags.Object | ts.TypeFlags.NonPrimitive));
}
for (const f of p.getSourceFiles().filter((f) => files.includes(f.fileName))) {
  function add(n, kind, extra = {}) {
    const pos = f.getLineAndCharacterOfPosition(n.getStart(f));
    sites.push({
      file: path.relative(root, f.fileName),
      line: pos.line + 1,
      column: pos.character + 1,
      kind,
      text: n.getText(f).replace(/\s+/g, " ").slice(0, 240),
      ...extra,
    });
  }
  function visit(n) {
    if (
      (ts.isStringLiteral(n) || ts.isNoSubstitutionTemplateLiteral(n)) &&
      ts.isPropertyAssignment(n.parent) &&
      n.parent.name.getText(f) === "text"
    ) {
      const helper = ts.createSourceFile(
        "emitted.js",
        n.text,
        ts.ScriptTarget.ESNext,
        true,
        ts.ScriptKind.JS,
      );
      const base = f.getLineAndCharacterOfPosition(n.getStart(f)).line + 1;
      function emitted(x) {
        let kind;
        if (ts.isForInStatement(x)) kind = "for...in";
        if (ts.isPropertyAccessExpression(x)) {
          const recv = x.expression.getText(helper),
            key = x.name.text;
          if (
            recv === "Object" &&
            [
              "keys",
              "entries",
              "assign",
              "getPrototypeOf",
              "setPrototypeOf",
              "create",
            ].includes(key)
          )
            kind = "Object." + key;
          if (["prototype", "__proto__", "hasOwnProperty"].includes(key))
            kind =
              key === "hasOwnProperty"
                ? "hasOwnProperty member"
                : "prototype access";
        }
        if (kind)
          helpers.push({
            file: path.relative(root, f.fileName),
            line:
              base +
              helper.getLineAndCharacterOfPosition(x.getStart(helper)).line,
            kind,
            text: x.getText(helper).replace(/\s+/g, " ").slice(0, 200),
          });
        ts.forEachChild(x, emitted);
      }
      emitted(helper);
    }

    if (ts.isForInStatement(n))
      add(n, "for...in", {
        receiver: c.typeToString(c.getTypeAtLocation(n.expression)),
      });
    if (ts.isPropertyAccessExpression(n) || ts.isElementAccessExpression(n)) {
      const key = ts.isPropertyAccessExpression(n)
        ? n.name.text
        : ts.isStringLiteral(n.argumentExpression)
          ? n.argumentExpression.text
          : null;
      const recv = n.expression.getText(f);
      if (["prototype", "__proto__"].includes(key))
        add(n, "prototype access", { receiver: recv });
      if (
        recv === "Object" &&
        [
          "keys",
          "entries",
          "assign",
          "getPrototypeOf",
          "setPrototypeOf",
          "create",
        ].includes(key)
      ) {
        const call = ts.isCallExpression(n.parent) && n.parent.expression === n;
        const arg = call ? n.parent.arguments[0] : undefined;
        let initializer;
        if (arg && ts.isIdentifier(arg)) {
          const symbol = c.getSymbolAtLocation(arg);
          const decl = symbol?.valueDeclaration;
          if (decl && ts.isVariableDeclaration(decl))
            initializer = decl.initializer;
        }
        add(n, "Object." + key, {
          call,
          arguments: call
            ? n.parent.arguments.map((a) =>
                c.typeToString(c.getTypeAtLocation(a)),
              )
            : [],
          argumentKind: arg ? ts.SyntaxKind[arg.kind] : undefined,
          initializerKind: initializer
            ? ts.SyntaxKind[initializer.kind]
            : undefined,
        });
      }
      if (key === "hasOwnProperty")
        add(n, "hasOwnProperty member", { receiver: recv });
      if (
        key === "stringify" &&
        recv === "JSON" &&
        ts.isCallExpression(n.parent)
      )
        add(n.parent, "JSON.stringify", {
          arguments: n.parent.arguments.map((a) =>
            c.typeToString(c.getTypeAtLocation(a)),
          ),
        });
      if (
        ["toString", "valueOf"].includes(key) &&
        ts.isCallExpression(n.parent) &&
        n.parent.expression === n
      )
        add(n.parent, key + " call", {
          receiver: c.typeToString(c.getTypeAtLocation(n.expression)),
        });
    }
    if (
      ts.isCallExpression(n) &&
      ts.isPropertyAccessExpression(n.expression) &&
      n.expression.name.text === "call" &&
      n.expression.expression.getText(f) === "hasOwnProperty"
    )
      add(n, "cached hasOwnProperty.call", {
        arguments: n.arguments.map((a) =>
          c.typeToString(c.getTypeAtLocation(a)),
        ),
      });
    if (ts.isTemplateSpan(n)) {
      const t = c.getTypeAtLocation(n.expression);
      if (possibleObject(t))
        add(n.expression, "template object conversion", {
          type: c.typeToString(t),
          definite: definiteObject(t),
        });
    }
    if (
      ts.isBinaryExpression(n) &&
      [ts.SyntaxKind.PlusToken, ts.SyntaxKind.PlusEqualsToken].includes(
        n.operatorToken.kind,
      )
    ) {
      const a = c.getTypeAtLocation(n.left),
        b = c.getTypeAtLocation(n.right);
      if (possibleObject(a) || possibleObject(b))
        add(n, "+ object conversion", {
          left: c.typeToString(a),
          right: c.typeToString(b),
          definite: definiteObject(a) || definiteObject(b),
        });
    }
    ts.forEachChild(n, visit);
  }
  visit(f);
}
sites.sort(
  (a, b) =>
    a.file.localeCompare(b.file) ||
    a.line - b.line ||
    a.column - b.column ||
    a.kind.localeCompare(b.kind),
);
const counts = {};
for (const s of sites) counts[s.kind] = (counts[s.kind] || 0) + 1;
const helperCounts = {};
for (const h of helpers) helperCounts[h.kind] = (helperCounts[h.kind] || 0) + 1;
const inventory = files.map((file) => ({
  file: path.relative(root, file),
  sha256: crypto
    .createHash("sha256")
    .update(fs.readFileSync(file))
    .digest("hex"),
}));
const result = {
  root,
  typescript: ts.version,
  files: files.length,
  inventory,
  diagnostics: ts.getPreEmitDiagnostics(p).length,
  counts,
  sites,
  helperCounts,
  helpers,
};
if (output) fs.writeFileSync(output, JSON.stringify(result, null, 2) + "\n");
else console.log(JSON.stringify(result, null, 2));
