"use strict";
// Usage: node build.cjs <srcDir> <libDir> <outDir>
// Builds tsc from <srcDir> without hereby or esbuild: every .ts that src/tsc/tsc.ts reaches through relative imports is
// transpiled alone with the API's transpileModule into <outDir>, keeping the `.js` specifiers as written, and
// <libDir>'s lib.*.d.ts are copied where the built sys.ts looks for its default libs.
//
// The output is CommonJS, not the ES modules the brief first asked for. The tree is CommonJS source written in import
// syntax (package.json has no "type", tsconfig-base says NodeNext, and upstream's own --bundle=false build runs tsc's
// CommonJS output directly). sys.ts reads __filename in getNodeSystem, eagerly (executingFilePath and
// isFileSystemCaseSensitive), and plain 6.0.3's isNodeLikeSystem also tests `typeof require`; neither exists in an ES
// module, so a {"type":"module"} build would have no sys and stop at its first `ts.sys` read.
//
// Each file is transpiled with no type information, as esbuild does upstream, so this works because the source is
// already written for it: const enums are emitted as objects (preserveConstEnums, which transpileModule's
// isolatedModules implies too) and read by property across files rather than inlined, and imports used only as types
// are dropped by syntax.
const fs = require("node:fs");
const path = require("node:path");
const { builtinModules } = require("node:module");
const common = require("./common.cjs");

const [srcDir, libDir, outDir] = process.argv.slice(2).map(a => a && path.resolve(a));
if (!srcDir || !libDir || !outDir) {
    console.error("usage: node build.cjs <srcDir> <libDir> <outDir>");
    process.exit(2);
}
const ts = common.loadTypeScript();
const compilerOptions = {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2022,
    // Upstream's esbuild build targets es2020 with no useDefineForClassFields, so class fields are assigned, not
    // defined; ES2022 would flip that default and change field semantics and object shapes.
    useDefineForClassFields: false,
    preserveConstEnums: true,
    alwaysStrict: true,
    newLine: ts.NewLineKind.LineFeed,
    sourceMap: false,
    declaration: false,
};

const entry = path.join(srcDir, "tsc/tsc.ts");
const queue = [entry];
const seen = new Set(queue);
const problems = [];
const warnings = new Set();
let written = 0;
while (queue.length) {
    const file = queue.shift();
    const text = fs.readFileSync(file, "utf8");
    for (const imported of ts.preProcessFile(text, true, false).importedFiles) {
        const specifier = imported.fileName;
        if (specifier.startsWith(".")) {
            const resolved = path.resolve(path.dirname(file), specifier.replace(/\.js$/, ".ts"));
            if (!fs.existsSync(resolved)) problems.push(`${path.relative(srcDir, file)}: ${specifier} has no .ts beside it`);
            else if (!seen.has(resolved)) {
                seen.add(resolved);
                queue.push(resolved);
            }
        }
        // A package name is only a warning: preProcessFile also reports `typeof import("x")` in a type, and sys.ts
        // names source-map-support that way behind a try. A static import of a package would fail the smoke run.
        else if (!builtinModules.includes(specifier.replace(/^node:/, ""))) {
            warnings.add(`${path.relative(srcDir, file)}: ${specifier} is a package, and the built tree has none`);
        }
    }
    const result = ts.transpileModule(text, { fileName: file, compilerOptions, reportDiagnostics: true });
    for (const d of result.diagnostics ?? []) {
        if (d.category === ts.DiagnosticCategory.Error) problems.push(`${path.relative(srcDir, file)}: ${ts.flattenDiagnosticMessageText(d.messageText, "\n")}`);
    }
    const out = path.join(outDir, path.relative(srcDir, file)).replace(/\.ts$/, ".js");
    fs.mkdirSync(path.dirname(out), { recursive: true });
    fs.writeFileSync(out, result.outputText);
    written++;
}
// An explicit CommonJS marker, so a package.json above the scratch directory can't make these ES modules.
fs.writeFileSync(path.join(outDir, "package.json"), "{\"type\":\"commonjs\"}\n");

// The built sys.js sees its own name end in sys.js and so takes its directory's parent, outDir, as the libs' home
// (upstream's unbundled road). The brief's place, beside the tsc entry, gets them too; it costs a copy.
const libs = fs.readdirSync(libDir).filter(n => /^lib\..*\.d\.ts$/.test(n) || n === "lib.d.ts");
for (const name of libs) {
    fs.copyFileSync(path.join(libDir, name), path.join(outDir, name));
    fs.copyFileSync(path.join(libDir, name), path.join(outDir, "tsc", name));
}
for (const w of warnings) console.log(`warning: ${w}`);
console.log(`built ${written} files from ${srcDir} into ${outDir}, ${libs.length} lib files from ${libDir}`);
if (!libs.length) problems.push(`${libDir} holds no lib.*.d.ts`);
if (problems.length) {
    for (const p of problems.slice(0, 40)) console.error(`  ${p}`);
    console.error(`${problems.length} problems`);
    process.exit(1);
}
