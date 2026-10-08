// Keeps allocation witnesses alive until measurement end. This is not a GC census.
exports.makeTracker=function(options={}){
 let stack=[],depth=0,records=[],audited=new Set(),paths=new Map();
 return {
  audit(node){if(!options.outsideOnly||!depth)audited.add(node);},
  enter(site,parser){stack.push({site,parser});if(parser)depth++;},
  leave(){const frame=stack.pop();if(!frame)throw Error('Unbalanced instrumentation');if(frame.parser)depth--;},
  record(node,allocation){
   if(options.outsideOnly&&depth)return node;
   const frames=stack.map(f=>f.site);const callerStack=options.captureOutside&&!depth?new Error().stack.split('\n').slice(2).map(line=>line.trim()):undefined;
   const key=allocation+'|'+frames.join(' > ')+(callerStack?'|'+callerStack.join(' > '):'');
   let callPath=paths.get(key);if(!callPath){callPath={allocation,frames,...(callerStack?{callerStack}: {})};paths.set(key,callPath);}
   records.push({node,origin:depth?'parser':'factory outside parsing',callPath});return node;
  },
  begin(){if(stack.length)throw Error('Open instrumented call');records=[];audited=new Set();},
  end(roots,ts){
   if(stack.length)throw Error('Open instrumented call');
   const syntax=new Set(),graph=new Set(),objects=new Set(),pending=[...roots];
   const visitSyntax=n=>{if(syntax.has(n))return;syntax.add(n);ts.forEachChild(n,visitSyntax);for(const j of n.jsDoc||[])visitSyntax(j);};
   for(const root of roots)visitSyntax(root);
   while(pending.length){const value=pending.pop();if(!value||typeof value!=='object'||objects.has(value))continue;objects.add(value);if(typeof value.kind==='number'&&typeof value.pos==='number'&&typeof value.end==='number')graph.add(value);for(const item of Object.values(value))if(item&&typeof item==='object')pending.push(item);}
   const registered=new Set(records.map(r=>r.node));
   if(registered.size!==records.length||registered.size!==audited.size||[...audited].some(n=>!registered.has(n)))throw Error(`Allocation coverage mismatch: ${registered.size} recorded, ${audited.size} audited`);
   const totals={auditedConstructorObjects:audited.size,created:records.length,parser:0,synthetic:0,noParent:0,sourceFiles:0,parentlessNonSourceFile:0,notSyntaxReachable:0,notGraphReachable:0,detachedLiteralUnion:0,detachedExcludingSourceFileRoots:0};const groups=new Map();
   for(const r of records){const n=r.node;const parser=r.origin==='parser';const noParent=n.parent===undefined;const sf=n.kind===ts.SyntaxKind.SourceFile;const unreachable=!graph.has(n);const detached=noParent||unreachable;
    totals[parser?'parser':'synthetic']++;totals.noParent+=+noParent;totals.sourceFiles+=+sf;totals.parentlessNonSourceFile+=+(noParent&&!sf);totals.notSyntaxReachable+=+!syntax.has(n);totals.notGraphReachable+=+unreachable;totals.detachedLiteralUnion+=+detached;totals.detachedExcludingSourceFileRoots+=+((noParent&&!sf)||unreachable);
    const key=r.origin+'|'+r.callPath.allocation+'|'+r.callPath.frames.join(' > ')+(r.callPath.callerStack?'|'+r.callPath.callerStack.join(' > '):'');let g=groups.get(key);if(!g){g={origin:r.origin,...r.callPath,created:0,noParent:0,notGraphReachable:0,detached:0,kinds:{}};groups.set(key,g);}g.created++;g.noParent+=+noParent;g.notGraphReachable+=+unreachable;g.detached+=+detached;const kind=ts.SyntaxKind[n.kind];g.kinds[kind]=(g.kinds[kind]||0)+1;
   }
   const report={totals,rootSourceFiles:roots.length,syntaxReachable:syntax.size,graphReachable:graph.size,paths:[...groups.values()]};records=[];audited=new Set();return report;
  }
 };
};
