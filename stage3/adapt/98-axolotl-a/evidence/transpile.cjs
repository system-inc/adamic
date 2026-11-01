// Transpiles each file from both trees with stock TypeScript 6.0.3, comments kept and removed, and writes the outputs.
const ts = require("typescript"), fs = require("fs"), path = require("path");
const [before, after, out] = process.argv.slice(2);
if (ts.version !== "6.0.3") throw new Error(ts.version);
for (const file of ["transformers/es2017.ts", "transformers/classFields.ts"]) for (const removeComments of [false, true]) for (const [label, tree] of [["before", before], ["after", after]]) {
  const name = path.join(tree, "src/compiler", file);
  const js = ts.transpileModule(fs.readFileSync(name, "utf8"), { fileName: name, compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } }).outputText;
  const target = path.join(out, `${path.basename(file, ".ts")}.${removeComments ? "nocomments" : "comments"}.${label}.js`);
  fs.writeFileSync(target, js);
}
