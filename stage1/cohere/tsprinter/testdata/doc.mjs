// Constructs docs with Prettier itself, independently of the port.

import {createRequire} from 'node:module';
import {readFileSync} from 'node:fs';
const require=createRequire(process.argv[2]+'/package.json');
if(require('prettier').version!=='3.9.6') throw Error('wrong Prettier version');
const {builders:b,printer}=require('prettier/doc');
const input=readFileSync(process.argv[3],"utf8");
{
  const results = JSON.parse(input).map(({ spec, options }) => {
    const ids = new Map(), groups = new Map();
    const id = (name) => { if (!name) return undefined; if (!ids.has(name)) ids.set(name, Symbol(name)); return ids.get(name); };
    const build = (node) => {
      const kids = () => (node.c || []).map(build);
      switch (node.k) {
        case "text": return node.v || "";
        case "line": return b.line;
        case "softline": return b.softline;
        case "hardline": return b.hardline;
        case "literalline": return b.literalline;
        case "hardlineWithoutBreakParent": return b.hardlineWithoutBreakParent;
        case "concat": return kids();
        case "indent": return b.indent(build(node.c[0]));
        case "alignWidth": return b.align(node.n || 0, build(node.c[0]));
        case "alignString": return b.align(node.v || "", build(node.c[0]));
        case "markAsRoot": return b.markAsRoot(build(node.c[0]));
        case "dedentToRoot": return b.dedentToRoot(build(node.c[0]));
        case "group": { const g = b.group(build(node.c[0]), { id: id(node.id), shouldBreak: !!node.f }); if (node.r) groups.set(node.r, g); return g; }
        case "sharedGroup": return groups.get(node.r);
        case "conditionalGroup": return b.conditionalGroup(kids(), { shouldBreak: !!node.f });
        case "fill": return b.fill(kids());
        case "ifBreak": return b.ifBreak(build(node.c[0]), build(node.c[1]), { groupId: id(node.id) });
        case "indentIfBreak": return b.indentIfBreak(build(node.c[0]), { groupId: id(node.id), negate: !!node.f });
        case "lineSuffix": return b.lineSuffix(build(node.c[0]));
        case "lineSuffixBoundary": return b.lineSuffixBoundary;
        case "breakParent": return b.breakParent;
        case "trim": return b.trim;
        case "label": return b.label(node.v, build(node.c[0]));
      }
      throw new Error("unknown spec kind " + node.k);
    };
    try {
      return { formatted: printer.printDocToString(build(spec), { ...options, endOfLine: "lf" }).formatted };
    } catch (error) {
      return { error: String(error) };
    }
  });
  const escape=text=>text.replaceAll('\\','\\\\').replaceAll('\n','\\n').replaceAll('\r','\\r').replaceAll('\t','\\t');
  for(const result of results) { if(result.error) throw Error(result.error); process.stdout.write('ok\t'+escape(result.formatted)+'\n'); }
}
