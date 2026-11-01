const ts = require("/home/user/adamic/stage3/api/node_modules/typescript");
const fs = require("fs"); const crypto = require("crypto");
const [before, after] = process.argv.slice(2).map(p => fs.readFileSync(p, "utf8"));
console.log("typescript", ts.version);
for (const removeComments of [false, true]) {
  const o = { fileName: "es2015.ts", compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } };
  const a = ts.transpileModule(before, o).outputText, b = ts.transpileModule(after, o).outputText;
  const h = s => crypto.createHash("sha256").update(s).digest("hex");
  console.log(`removeComments=${removeComments} before=${h(a)} after=${h(b)} bytes=${a.length}/${b.length} identical=${a === b}`);
}
