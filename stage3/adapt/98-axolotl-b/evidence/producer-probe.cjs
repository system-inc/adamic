const ts = require("/home/user/adamic/stage3/api/node_modules/typescript");
const cases = ["for (const [,] of xs) {}", "for (const [] of xs) {}", "for (const {} of xs) {}", "for (var [,,] of xs) {}", "for (let {a: [,]} of xs) {}", "f(...a)", "new C(...a)", "[...a]", "class B extends A {}; @d class C extends B { static x = () => C; }"];
for (const c of cases) for (const dl of [false, true]) {
  const out = ts.transpileModule(c, { compilerOptions: { target: ts.ScriptTarget.ES5, downlevelIteration: dl, experimentalDecorators: true, ignoreDeprecations: "6.0" } }).outputText;
  console.log(JSON.stringify(c), dl, JSON.stringify(out.split("\n").slice(0,3).join(" | ")));
}
