import pathlib,json,difflib,subprocess
root=pathlib.Path('/workspace/adamic');out=root/'review/test-defend/internal-flow'
recipes=[
('D1','internal/flow/build.go','\tb.statements(statement.Update)\n','','Drop loop update statements; timsort repeatedly increments index and shape at loop updates'),
('D2','internal/flow/build.go','\tb.enter(head)\n\tb.current = after','\tb.enter(after)\n\tb.current = after','Change the for-of back-edge target to exit; timsort traverses its length array repeatedly'),
('D3','internal/flow/ssa.go','for index := range b.function.Params {','for index := 0; index+1 < len(b.function.Params); index++ {','Off-by-one parameter renaming; timsort comparators and numbers each use two parameters'),
('D4','internal/flow/ssa.go','\t\t\tremaining--','\t\t\tremaining -= 2','Off-by-one decrement of unsealed predecessors; nested loops and branch joins need late phis'),
('D5','internal/flow/build.go','b.current.Terminal = &If{Consequent: body.Id, Alternate: after.Id}','b.current.Terminal = &If{Consequent: body.Id, Alternate: body.Id}','Change loop exit edge to body; timsort exercises both empty and long counted loops'),
('D6','internal/flow/ranges.go','\t\t\t\t\t\tend:        instruction.Order + 1,\n\t\t\t\t\t\ttransitive: true,','\t\t\t\t\t\tend:        instruction.Order,\n\t\t\t\t\t\ttransitive: true,','Off-by-one transitive mutation end; in-place array sort and push need last mutation included'),
('D7','internal/flow/ranges.go','\t\tfor _, pending := range pendingPhis[block.Id] {\n\t\t\tstate.assign(pending.index, pending.from, pending.into)\n\t\t}\n','','Drop whole deferred loop-phi alias assignment loop; mutations through loop-carried arrays')]
plan=[]
for mid,file,old,new,aim in recipes:
 original=(root/file).read_text();assert original.count(old)==1,(mid,original.count(old));change=original.replace(old,new)
 line=original[:original.index(old)].count('\n')+1
 (out/(mid+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),change.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 plan.append(dict(mutant=mid,file=file,file_line=file+':'+str(line),old=old,new=new,aim=aim))
(out/'plan.json').write_text(json.dumps(plan,indent=2))
