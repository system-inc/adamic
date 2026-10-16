// Plan indivisible source statements with stock TypeScript, independently of Adamic.
const fs = require("node:fs"), path = require("node:path"), crypto = require("node:crypto");
const ts = require("typescript");
if (ts.version !== "6.0.3") throw Error("requires TypeScript 6.0.3");
const [filename, countText, output] = process.argv.slice(2);
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
function atom(node, parent) {
    const start = offsets[Math.max(0, node.pos)], end = offsets[node.end];
    // SyntaxKind has reverse-name aliases; prefer the forward named statement kind.
    const kind = Object.keys(ts.SyntaxKind).find(k => !/^\d+$/.test(k) && ts.SyntaxKind[k] === node.kind && !/^(First|Last|Count)/.test(k));
    atoms.push({id:`${start}:${end}:${kind}`, start, end, kind, parent, bytes:end-start,
                name:node.name && ts.isIdentifier(node.name) ? node.name.text : ""});
}
for (const node of source.statements) {
    if (ts.isFunctionDeclaration(node) && node.name?.text === "createTypeChecker" && node.body) {
        for (const child of node.body.statements) atom(child, "createTypeChecker");
    } else atom(node, "file");
}
const pieces = Array.from({length:Math.min(requested, atoms.length)}, (_,id) => ({id, atoms:[], bytes:0}));
// Largest-first scheduling minimizes the largest bucket subject to indivisible atoms.
for (const node of [...atoms].sort((a,b) => b.bytes-a.bytes || a.start-b.start)) {
    const piece = [...pieces].sort((a,b) => a.bytes-b.bytes || a.id-b.id)[0];
    piece.atoms.push(node.id); piece.bytes += node.bytes;
}
const order = new Map(atoms.map((a,i)=>[a.id,i]));
for (const piece of pieces) piece.atoms.sort((a,b)=>order.get(a)-order.get(b));
fs.writeFileSync(output, JSON.stringify({version:1, typescript:ts.version, file, sha256:crypto.createHash("sha256").update(bytes).digest("hex"),
    source_bytes:bytes.length, requested_pieces:requested, atom_definition:"top-level statements; createTypeChecker body statements replace their wrapper", atoms,pieces},null,2)+"\n");
console.log(JSON.stringify({file, atoms:atoms.length, pieces:pieces.length, sizes:pieces.map(p=>p.bytes)}));
