from pathlib import Path
import subprocess,json,tempfile
root=Path(__file__).resolve().parents[2];source=root/'internal/lower/discriminant_writes.go';original=source.read_text();out=Path(tempfile.mkdtemp(prefix='adamic-discriminant-mutants-'))
mutants={
 'skip_check':(original.replace('func (l *lowering) refuseDiscriminantWrites(module *ast.SourceFile) error {','func (l *lowering) refuseDiscriminantWrites(module *ast.SourceFile) error {\n return nil\n').replace('func (l *lowering) checkDiscriminantConstruction(modules []*ast.SourceFile) error {','func (l *lowering) checkDiscriminantConstruction(modules []*ast.SourceFile) error {\n return nil\n'),'TestDiscriminantWrites/union'),
 'first_member':(original.replace('for _, member := range l.discriminantMembers(fields, holder, name) {','for index, member := range l.discriminantMembers(fields, holder, name) {\n if index>0 {break}'),'TestDiscriminantWrites/union'),
 'every_holder_fresh':(original.replace('func (l *lowering) checkDiscriminantConstruction(modules []*ast.SourceFile) error {','func (l *lowering) checkDiscriminantConstruction(modules []*ast.SourceFile) error {\n return nil\n'),'TestDiscriminantWrites/this_escapes_constructor'),
 'miss_compound':(original.replace('func (l *lowering) unsafeDiscriminantWrite(fields []discriminantField, holder, value *checker.Type, name string) bool {','func (l *lowering) unsafeDiscriminantWrite(fields []discriminantField, holder, value *checker.Type, name string) bool {\n if value!=nil && value.Flags()&(checker.TypeFlagsString|checker.TypeFlagsNumber)!=0 { return false }'),'TestDiscriminantWrites/compound'),
 'miss_base_view':(original.replace('func (l *lowering) unsafeDiscriminantWrite(fields []discriminantField, holder, value *checker.Type, name string) bool {','func (l *lowering) unsafeDiscriminantWrite(fields []discriminantField, holder, value *checker.Type, name string) bool {\n if holder!=nil && holder.Flags()&checker.TypeFlagsUnion==0 { return false }'),'TestDiscriminantWrites/base_interface'),
 'constructor_update':(original.replace('func (l *lowering) discriminantConstructionUpdate(target, receiver *ast.Node, name string) bool {','func (l *lowering) discriminantConstructionUpdate(target, receiver *ast.Node, name string) bool {\n return false\n'),'TestDiscriminantCompoundConstruction'),
}
for name,(content,test) in mutants.items():
 target=out/(name+'.go');target.write_text(content)
 overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(source):str(target)}}))
 log=Path('/tmp/discriminant-ruled-mutant-'+name+'.log')
 with log.open('w') as f:
  r=subprocess.run(['go','test','-overlay',str(overlay),'./internal/lower','-run','^'+test+'$','-count=1','-v'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
 print(name,r.returncode,str(log),flush=True)
 if r.returncode==0:raise SystemExit('survived mutant '+name)
 text=log.read_text()
 if 'want discriminant refusal, got <nil>' not in text and 'want constructor tag update refused, got <nil>' not in text:raise SystemExit('wrong killer '+name+': '+text)
