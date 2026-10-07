#!/usr/bin/env python3
"""Add read queries to lane 3's adapter. All solving stays in its shared graph."""
import pathlib, subprocess, sys
repo=pathlib.Path.cwd(); scratch=pathlib.Path(sys.argv[1]);scratch.mkdir(parents=True,exist_ok=True)
assert subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()=='c1f4c5a70bd7d98fd543f4eb7ff96191be4a338b'
source=(repo/'stage3/shape-conformance/latent/lower.go.txt').read_text()
insert='''
 readQueries := map[string]ir.Expression{}
 for _, file := range program.Files() {
  var visit ast.Visitor
  visit = func(node *ast.Node) bool {
   var receiver *ast.Node
   switch node.Kind {
   case ast.KindPropertyAccessExpression: receiver = node.AsPropertyAccessExpression().Expression
   case ast.KindElementAccessExpression: receiver = node.AsElementAccessExpression().Expression
   }
   if receiver != nil && !ast.IsPartOfTypeNode(node) {
    key := latentShapeKey(m.file(node), scanner.GetTokenPosOfNode(node, file, false), node.End())
    if m.Diagnosed[node] {
     readQueries[key] = latentShapeUnknown{Category:"diagnosed body", Reason:"receiver in diagnosed body; retain read check"}
    } else {
     m.Current = -1
     for parent := node.Parent; parent != nil; parent = parent.Parent {
      if function, ok := m.Functions[parent]; ok { m.Current = function.Index; break }
     }
     readQueries[key] = m.expression(receiver)
    }
   }
   node.ForEachChild(visit)
   return false
  }
  file.AsNode().ForEachChild(visit)
 }
'''
marker='\tm.Current = -1\n\tassignAllocationSites(m.IR)'
assert source.count(marker)==1
source=source.replace(marker,insert+marker)
query='''
 // Unknown cast frontiers cannot prove disjoint receivers non-viewed: an
 // unmodeled property load could carry a known allocation to that cast.
 viewed := map[int]bool{}
 viewUnknown := false
 for _, site := range sites {
  for _, id := range site.AllocationSites { viewed[id] = true }
  if site.Outcome == "unknown" { viewUnknown = true }
 }
 keys := []string{}
 for key := range readQueries { keys = append(keys,key) }
 sort.Strings(keys)
 readReceivers := []map[string]any{}
 cache := map[int]ir.AllocationSet{}
 for _, key := range keys {
  expression := readQueries[key]
  var set ir.AllocationSet
  if local, ok := expression.(ir.Read); ok {
   var exists bool
   set, exists = cache[local.Local]
   if !exists { set = graph.ReachingAllocations(expression); cache[local.Local] = set }
  } else { set = graph.ReachingAllocations(expression) }
  overlap := []int{}
  for _, id := range set.Sites { if viewed[id] { overlap = append(overlap,id) } }
  demand := set.Unknown || viewUnknown || len(overlap) != 0
  readReceivers = append(readReceivers,map[string]any{"key":key,"allocation_sites":set.Sites,"graph_unknown":set.Unknown,"graph_reasons":set.Reasons,"viewed_allocation_overlap":overlap,"unknown_cast_frontier":viewUnknown,"retain_check":demand})
 }
'''
marker='\treturn map[string]any{"measurement": latentShapeLabel'
assert source.count(marker)==1
source=source.replace(marker,query+marker).replace('"allocation_shapes": len(allocations),','"read_receivers": readReceivers, "read_receiver_count":len(readReceivers), "allocation_shapes": len(allocations),')
hook=scratch/'read-lower.go.txt';hook.write_text(source)
subprocess.run([sys.executable,str(repo/'stage3/shape-conformance/latent/make-overlay.py'),str(scratch),str(hook)],check=True)
