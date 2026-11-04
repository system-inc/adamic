// Usage: node select-and-instrument.cjs <tsRepo> <instrumentedCopyDir> <outDir>
// Selects slice-4 cast sites in <tsRepo>/src/compiler with the TS 6.0.3 API, writes sites.csv/reads.csv,
// and rewrites the files in <instrumentedCopyDir> so each cast execution and each view read bumps a counter.
const path = require("path"), fs = require("fs");
const [repo, copy, out] = process.argv.slice(2);
const ts = require(path.join(repo, "built/local/typescript.js"));
const cfgPath = path.join(repo, "src/compiler/tsconfig.json");
const cfg = ts.getParsedCommandLineOfConfigFile(cfgPath, {}, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: d => { throw new Error(ts.flattenDiagnosticMessageText(d.messageText, "\n")); } });
const program = ts.createProgram(cfg.fileNames, cfg.options);
const checker = program.getTypeChecker();
const SK = ts.SyntaxKind;

function stripParens(n) { while (n.kind === SK.ParenthesizedExpression) n = n.expression; return n; }
function isInterface(t) { const s = t.symbol; return !!s && (s.flags & ts.SymbolFlags.Interface) !== 0 && !(t.flags & ts.TypeFlags.Union); }
function derivesFromNode(t) { return !!checker.getPropertyOfType(t, "kind") && !!checker.getPropertyOfType(t, "pos") && !!checker.getPropertyOfType(t, "parent"); }
function enumLiteralValues(t) { // null when not wholly enum literals
  if (checker.typeToString(t) === "SyntaxKind") return null; // the open kind itself is no tag
  const members = t.flags & ts.TypeFlags.Union ? t.types : [t];
  const vals = [];
  for (const m of members) { if (!(m.flags & ts.TypeFlags.EnumLiteral) || !(m.flags & ts.TypeFlags.NumberLiteral)) return null; vals.push(m.value); }
  return vals;
}
function kindType(t) { const p = checker.getPropertyOfType(t, "kind"); return p ? checker.getTypeOfSymbol(p) : undefined; }

const sites = [];   // per-file selected cast nodes
const allAsCount = { total: 0, nodeSource: 0 };
const interfaceTargetKinds = new Map(); // target type name -> Set of kind values, for uniqueness
const pending = [];
for (const sf of program.getSourceFiles()) {
  if (sf.isDeclarationFile || !sf.fileName.startsWith(path.join(repo, "src/compiler") + path.sep)) continue;
  (function walk(n) {
    if (n.kind === SK.AsExpression || n.kind === SK.TypeAssertionExpression) {
      allAsCount.total++;
      const typeNode = n.type;
      if (typeNode.kind === SK.TypeReference && typeNode.typeName.getText() === "const") { /* as const */ }
      else {
        const S = checker.getTypeAtLocation(n.expression), T = checker.getTypeAtLocation(n);
        const sk = S.flags & ts.TypeFlags.Union ? undefined : kindType(S);
        const sourceOk = !(S.flags & ts.TypeFlags.Union) && isInterface(S) && derivesFromNode(S) && sk && checker.typeToString(sk) === "SyntaxKind";
        if (sourceOk) {
          allAsCount.nodeSource++;
          const members = T.flags & ts.TypeFlags.Union ? T.types : [T];
          const shape = T.flags & ts.TypeFlags.Union ? "union member" : "interface downcast";
          let ok = members.every(m => isInterface(m) && derivesFromNode(m));
          let tags = [];
          if (ok) for (const m of members) { const v = enumLiteralValues(kindType(m) ?? m); if (!v) { ok = false; break; } tags.push(...v); }
          if (ok) {
            const { line, character } = sf.getLineAndCharacterOfPosition(n.getStart());
            pending.push({ sf, n, S: checker.typeToString(S), T: checker.typeToString(T), shape, tags });
            for (const m of members) { const k = m.symbol.name; if (!interfaceTargetKinds.has(k)) interfaceTargetKinds.set(k, new Set()); for (const v of enumLiteralValues(kindType(m))) interfaceTargetKinds.get(k).add(v); }
          }
        }
      }
    }
    ts.forEachChild(n, walk);
  })(sf);
}
// unique: the target's kind values are claimed by no other selected target type
const claim = new Map(); for (const [name, vs] of interfaceTargetKinds) for (const v of vs) claim.set(v, (claim.get(v) || new Set()).add(name));
const selected = pending.map((p, i) => ({ ...p, id: i, rel: path.relative(repo, p.sf.fileName), unique: p.tags.every(v => claim.get(v).size === 1) }));
const byNode = new Map(selected.map(s => [s.n, s]));

// reads: property accesses whose receiver is the cast, or a never-reassigned const/let bound to it
const edits = new Map(); // file -> [{pos,end,text}] (insertions have pos==end)
const addEdit = (sf, pos, end, text) => { const k = sf.fileName; if (!edits.has(k)) edits.set(k, []); edits.get(k).push({ pos, end, text }); };
const bindings = new Map(); // symbol -> site
for (const s of selected) {
  let p = s.n.parent; while (p.kind === SK.ParenthesizedExpression) p = p.parent;
  if (ts.isVariableDeclaration(p) && ts.isIdentifier(p.name) && stripParens(p.initializer) === s.n) {
    const list = p.parent; const sym = checker.getSymbolAtLocation(p.name);
    if (sym && (list.flags & (ts.NodeFlags.Const | ts.NodeFlags.Let))) bindings.set(sym, { s, decl: p, reassigned: false });
  }
}
const isAssignOp = k => k >= SK.FirstAssignment && k <= SK.LastAssignment;
function isWriteTarget(n) { // n is the property access / identifier; true when only assigned (or incremented/deleted)
  let c = n, p = n.parent;
  if (ts.isBinaryExpression(p) && p.left === c && isAssignOp(p.operatorToken.kind)) return p.operatorToken.kind === SK.EqualsToken ? "write" : "readwrite";
  if ((ts.isPrefixUnaryExpression(p) || ts.isPostfixUnaryExpression(p)) && (p.operator === SK.PlusPlusToken || p.operator === SK.MinusMinusToken)) return "readwrite";
  if (ts.isDeleteExpression(p)) return "write";
  return null;
}
// first pass: reassignment detection
for (const sf of program.getSourceFiles()) {
  if (sf.isDeclarationFile || !sf.fileName.startsWith(path.join(repo, "src/compiler") + path.sep)) continue;
  (function walk(n) {
    if (ts.isIdentifier(n) && !ts.isVariableDeclaration(n.parent)) { const sym = checker.getSymbolAtLocation(n); const b = sym && bindings.get(sym); if (b && isWriteTarget(n)) b.reassigned = true; }
    ts.forEachChild(n, walk);
  })(sf);
}
const reads = []; // {site, kind: direct|binding|destructure, access: read|write|readwrite, where}
let instrumentedFiles = new Set();
for (const sf of program.getSourceFiles()) {
  if (sf.isDeclarationFile || !sf.fileName.startsWith(path.join(repo, "src/compiler") + path.sep)) continue;
  (function walk(n) {
    // direct and binding receivers
    if ((ts.isPropertyAccessExpression(n) || ts.isElementAccessExpression(n))) {
      const recv = n.expression; let site, via;
      const inner = stripParens(recv);
      if (byNode.has(inner)) { site = byNode.get(inner); via = "direct"; }
      else if (ts.isIdentifier(recv)) { const sym = checker.getSymbolAtLocation(recv); const b = sym && bindings.get(sym); if (b && !b.reassigned) { site = b.s; via = "binding"; } }
      if (site && ts.isPropertyAccessExpression(n)) {
        const w = isWriteTarget(n) || "read";
        reads.push({ id: site.id, via, access: w, file: site.rel, line: sf.getLineAndCharacterOfPosition(n.getStart()).line + 1, name: n.name.text });
        if (w !== "write") { // a plain write is not a read; a compound write counts as one
          addEdit(sf, recv.getStart(), recv.getStart(), `(__s4r[${site.id}]++,`);
          addEdit(sf, recv.end, recv.end, `)`);
        }
      }
    }
    // destructuring a cast / a bound const: count one read per element
    if (ts.isVariableDeclaration(n) && ts.isObjectBindingPattern(n.name) && n.initializer) {
      const inner = stripParens(n.initializer); let site, via;
      if (byNode.has(inner)) { site = byNode.get(inner); via = "destructure-direct"; }
      else if (ts.isIdentifier(inner)) { const sym = checker.getSymbolAtLocation(inner); const b = sym && bindings.get(sym); if (b && !b.reassigned) { site = b.s; via = "destructure-binding"; } }
      if (site) {
        const k = n.name.elements.length;
        for (const e of n.name.elements) reads.push({ id: site.id, via, access: "read", file: site.rel, line: sf.getLineAndCharacterOfPosition(e.getStart()).line + 1, name: (e.propertyName || e.name).getText() });
        addEdit(sf, n.initializer.getStart(), n.initializer.getStart(), `(__s4r[${site.id}]+=${k},`);
        addEdit(sf, n.initializer.end, n.initializer.end, `)`);
      }
    }
    ts.forEachChild(n, walk);
  })(sf);
}
for (const s of selected) { // cast execution counter
  addEdit(s.sf, s.n.getStart(), s.n.getStart(), `(__s4c[${s.id}]++,`);
  addEdit(s.sf, s.n.end, s.n.end, `)`);
}
fs.mkdirSync(out, { recursive: true });
const csv = (rows, cols) => cols.join(",") + "\n" + rows.map(r => cols.map(c => JSON.stringify(r[c] ?? "")).join(",")).join("\n") + "\n";
const siteRows = selected.map(s => ({ id: s.id, file: s.rel, line: s.sf.getLineAndCharacterOfPosition(s.n.getStart()).line + 1, shape: s.shape, source: s.S, target: s.T, unique: s.unique, text: s.n.getText().replace(/\s+/g, " ").slice(0, 80) }));
fs.writeFileSync(path.join(out, "sites.csv"), csv(siteRows, ["id", "file", "line", "shape", "source", "target", "unique", "text"]));
fs.writeFileSync(path.join(out, "read-sites.csv"), csv(reads, ["id", "via", "access", "file", "line", "name"]));
fs.writeFileSync(path.join(out, "selection.json"), JSON.stringify({ allAs: allAsCount, selected: selected.length, uniqueSelected: selected.filter(s => s.unique).length, byShape: Object.fromEntries(["interface downcast", "union member"].map(k => [k, selected.filter(s => s.shape === k).length])), staticReadSites: reads.length, readsByVia: reads.reduce((a, r) => (a[r.via + "/" + r.access] = (a[r.via + "/" + r.access] || 0) + 1, a), {}) }, null, 1));
// apply edits to the copy (descending), add header
for (const [file, es] of edits) {
  let text = fs.readFileSync(file, "utf8");
  es.sort((a, b) => b.pos - a.pos || b.end - a.end);
  // for equal positions the later-listed insertion must go first when closing; sort stable by insertion order reversed
  for (const e of es) text = text.slice(0, e.pos) + e.text + text.slice(e.end);
  const header = `var __s4c: Float64Array = ((globalThis as any).__s4c ??= new Float64Array(${selected.length})), __s4r: Float64Array = ((globalThis as any).__s4r ??= new Float64Array(${selected.length}));\n`;
  fs.writeFileSync(path.join(copy, path.relative(repo, file)), header + text);
}
console.log(fs.readFileSync(path.join(out, "selection.json"), "utf8"));
