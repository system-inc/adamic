from pathlib import Path
import shutil,subprocess
root=Path('/tmp/trees-profile'); dst=root/'runtime-diagnostic';shutil.copytree(root/'runtime-before',dst,dirs_exist_ok=True)
p=dst/'heap.c';s=p.read_text(); pos=s.index('static _Thread_local chunk *giving');s=s[:pos]+'''static size_t probe_take, probe_reuse, probe_fresh, probe_new, probe_spare, probe_drain, probe_empty, probe_children, probe_go, probe_null;
'''+s[pos:]
s=s.replace('chunk *each = spares;','probe_new++;\n\tchunk *each = spares;').replace('spares = each->next;','probe_spare++; spares = each->next;')
s=s.replace('static void *take(size_t class, uint32_t *number) {','static void *take(size_t class, uint32_t *number) { probe_take++;').replace('slot = each->free;','probe_reuse++; slot = each->free;').replace('slot = each->fresh;','probe_fresh++; slot = each->fresh;')
s=s.replace('static void drain_remote(chunk *each) {','static void drain_remote(chunk *each) { probe_drain++; if (atomic_load_explicit(&each->remote, memory_order_relaxed) == NULL) { probe_empty++; }')
s=s.replace('static void let_go(void *value) {','static void let_go(void *value) { probe_go++; if (value == NULL) { probe_null++; }')
s=s.replace('void adamic_heap_free_children(void *value, void (*let_go)(void *)) {','void adamic_heap_free_children(void *value, void (*let_go)(void *)) { probe_children++;')
s+='''
#include <stdio.h>
__attribute__((destructor)) static void probe_report(void) {
 fprintf(stderr,"slab take %zu reused %zu fresh %zu new_chunks %zu spare_chunks %zu drain %zu empty %zu children %zu let_go %zu null %zu chunk_header %zu\\n",probe_take,probe_reuse,probe_fresh,probe_new,probe_spare,probe_drain,probe_empty,probe_children,probe_go,probe_null,(size_t)FIRST_SLOT);
}
''';p.write_text(s)
p=dst/'region.c';s=p.read_text();s=s.replace('struct adamic_region_block {','static size_t probe_blocks, probe_bytes, probe_ends, probe_walks;\nstruct adamic_region_block {');s=s.replace('adamic_region_block *fresh = malloc','probe_blocks++; probe_bytes += capacity;\n\t\tadamic_region_block *fresh = malloc');s=s.replace('void adamic_region_end(adamic_region *region) {','void adamic_region_end(adamic_region *region) { probe_ends++;');s=s.replace('if (region->holds_outside) {','if (region->holds_outside) { probe_walks++;');s+='''
#include <stdio.h>
__attribute__((destructor)) static void probe_region_report(void) {
 fprintf(stderr,"region blocks %zu capacity_bytes %zu ends %zu child_walks %zu object_size %zu aligned_size %zu\\n",probe_blocks,probe_bytes,probe_ends,probe_walks,sizeof(adamic_object)+2*sizeof(adamic_value),object_size(2));
}
''';p.write_text(s)
flags=['-std=c11','-O2','-flto=thin','-fuse-ld=lld','-pthread','-ffp-contract=off','-fno-optimize-sibling-calls']
subprocess.run(['clang',*flags,'-I',str(dst),str(root/'trees.c'),*[str(x) for x in sorted(dst.glob('*.c'))],'-lm','-o',str(root/'diagnostic')],check=True)
p=subprocess.run([str(root/'diagnostic')],capture_output=True,check=True);assert p.stdout==(root/'counted.stdout').read_bytes();(root/'diagnostic.txt').write_bytes(p.stderr);print(p.stderr.decode())
