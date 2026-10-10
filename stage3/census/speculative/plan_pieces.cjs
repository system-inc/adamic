// Plan indivisible source statements with stock TypeScript, independently of Adamic.
const fs = require("node:fs"), path = require("node:path"), crypto = require("node:crypto");
const ts = require("typescript");
if (ts.version !== "6.0.3") throw Error("requires TypeScript 6.0.3");
const [filename, countText, output, ...options] = process.argv.slice(2);
const thresholdIndex = options.indexOf("--threshold"), rootIndex = options.indexOf("--root");
const threshold = thresholdIndex < 0 ? null : Number(options[thresholdIndex + 1]);
const rootName = rootIndex < 0 ? null : options[rootIndex + 1];
if (rootName && threshold === null) throw Error("root selection requires --threshold");
if (threshold !== null && (!Number.isInteger(threshold) || threshold < 1)) throw Error("invalid byte threshold");
const splitFunctions = [], splitContainers = [];
const file = path.resolve(filename), requested = Number(countText);
if (!Number.isInteger(requested) || requested < 1 || !output) throw Error("usage: plan_pieces.cjs FILE N OUTPUT");
const bytes = fs.readFileSync(file), text = bytes.toString("utf8");
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const offsets = new Uint32Array(text.length + 1);
let units = 0, byteOffset = 0;
for (const character of text) {
    if (character.length === 2) offsets[units + 1] = byteOffset + 3;
    units += character.length; byteOffset += Buffer.byteLength(character); offsets[units] = byteOffset;
}
const atoms = [];
function atom(node, parent, weight = null) {
    const start = offsets[Math.max(0, node.pos)], end = offsets[node.end];
    // SyntaxKind has reverse-name aliases; prefer the forward named statement kind.
    const kind = Object.keys(ts.SyntaxKind).find(k => !/^\d+$/.test(k) && ts.SyntaxKind[k] === node.kind && !/^(First|Last|Count)/.test(k));
    atoms.push({id:`${start}:${end}:${kind}`, start, end, kind, parent, bytes:weight ?? end-start,
                name:node.name && ts.isIdentifier(node.name) ? node.name.text : ""});
}
function describe(node) {
    return {start:offsets[Math.max(0,node.pos)], end:offsets[node.end], name:node.name?.text || ""};
}
function cut(node, parent) {
    const size = offsets[node.end] - offsets[Math.max(0,node.pos)];
    if (threshold === null || size <= threshold) { atom(node,parent); return; }
    if (ts.isFunctionDeclaration(node) && node.body) {
        const children = node.body.statements.filter(ts.isFunctionDeclaration);
        const own = size - children.reduce((n,c)=>n+offsets[c.end]-offsets[Math.max(0,c.pos)],0);
        if (!children.length || own > threshold) throw Error(`unsplittable own statements: ${node.name?.text} ${own} bytes`);
        splitFunctions.push(describe(node));
        atom(node,parent,own);
        for (const child of children) cut(child,node.name?.text || "function");
    } else if (ts.isModuleDeclaration(node) && node.body && ts.isModuleBlock(node.body)) {
        const children = node.body.statements;
        const own = size - children.reduce((n,c)=>n+offsets[c.end]-offsets[Math.max(0,c.pos)],0);
        if (own > threshold) throw Error("unsplittable namespace envelope");
        splitContainers.push(describe(node)); atom(node,parent,own);
        for (const child of children) cut(child,node.name?.text || "namespace");
    } else throw Error(`unsplittable ${ts.SyntaxKind[node.kind]}: ${size} bytes`);
}
if (rootName) {
    const matches=[];
    function find(node) { if(ts.isFunctionDeclaration(node) && node.name?.text===rootName) matches.push(node); ts.forEachChild(node,find); }
    find(source);
    if(matches.length!==1) throw Error("root function must be unique");
    cut(matches[0],"root");
} else for (const node of source.statements) {
    if (ts.isFunctionDeclaration(node) && node.name?.text === "createTypeChecker" && node.body) {
        for (const child of node.body.statements) cut(child, "createTypeChecker");
    } else cut(node, "file");
}
const pieces = Array.from({length:Math.min(requested, atoms.length)}, (_,id) => ({id, atoms:[], bytes:0}));
// Largest-first scheduling minimizes the largest bucket subject to indivisible atoms.
for (const node of [...atoms].sort((a,b) => b.bytes-a.bytes || a.start-b.start)) {
    let piece = [...pieces].sort((a,b) => a.bytes-b.bytes || a.id-b.id).find(p => threshold === null || p.bytes + node.bytes <= threshold);
    if (!piece) { piece={id:pieces.length,atoms:[],bytes:0}; pieces.push(piece); }
    piece.atoms.push(node.id); piece.bytes += node.bytes;
}
const order = new Map(atoms.map((a,i)=>[a.id,i]));
for (let i=pieces.length-1;i>=0;i--) if (!pieces[i].atoms.length) pieces.splice(i,1);
for (let i=0;i<pieces.length;i++) pieces[i].id=i;
for (const piece of pieces) piece.atoms.sort((a,b)=>order.get(a)-order.get(b));
// The legacy atom_definition labels the starting inventory; root_name and the
// explicit recursive cuts describe its version-2 refinement.
fs.writeFileSync(output, JSON.stringify({version:threshold === null ? 1 : 2, threshold_bytes:threshold, root_name:rootName, split_functions:splitFunctions, split_containers:splitContainers, typescript:ts.version, file, sha256:crypto.createHash("sha256").update(bytes).digest("hex"),
    source_bytes:bytes.length, requested_pieces:requested, atom_definition:"top-level statements; createTypeChecker body statements replace their wrapper", atoms,pieces},null,2)+"\n");
console.log(JSON.stringify({file, atoms:atoms.length, pieces:pieces.length, sizes:pieces.map(p=>p.bytes)}));
