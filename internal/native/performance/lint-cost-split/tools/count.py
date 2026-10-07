from pathlib import Path
import shutil,re,json
src=Path('/workspace/lint-cost-baseline');dst=Path('/workspace/lint-cost-instrumented');dst.mkdir(exist_ok=True)
for p in src.iterdir():
 if p.suffix in ['.c','.h'] or p.name in ['oracle','compiler.txt']:shutil.copy2(p,dst/p.name)
rules=json.loads(Path('/workspace/lint-cost-rule-kinds.json').read_text());s=(dst/'main.c').read_text()
pre='\n#include <stdio.h>\n#include <inttypes.h>\nstatic uint64_t cost_nodes = 0;\nstatic uint64_t cost_eligible[10] = {0};\n'
for i,kind in enumerate(dict.fromkeys(k for _,kinds in rules for k in kinds)):pre+=f'static adamic_string cost_kind_{i} = ADAMIC_STRING("{kind}");\n'
kind_ids={kind:i for i,kind in enumerate(dict.fromkeys(k for _,kinds in rules for k in kinds))};s=s.replace('#include "adamic.h"','#include "adamic.h"'+pre,1)
anchor='static void adamic_function_221_visit(adamic_object * adamic_local_1076_context, double adamic_local_1077_index) {';assert s.count(anchor)==1
body='\n cost_nodes++;\n adamic_object *cost_node = adamic_function_45_Context_node(adamic_local_1076_context, adamic_local_1077_index);\n adamic_string *cost_kind = (adamic_string *)cost_node->slots[0].reference;\n'
for i,(_,kinds) in enumerate(rules):body+=' if('+ ' || '.join(f'adamic_string_equal(cost_kind, &cost_kind_{kind_ids[k]})' for k in kinds)+f') cost_eligible[{i}]++;\n'
body+=' adamic_release(cost_node);\n';s=s.replace(anchor,anchor+body)
# Main ends with a zero return; counters are printed after the same complete count-only walk.
where=s.rfind('return 0;');assert where>=0
printing='fprintf(stderr,"nodes %" PRIu64 "\\n",cost_nodes);\n'
for i,(name,_) in enumerate(rules):printing+=f'fprintf(stderr,"{name} %" PRIu64 "\\n",cost_eligible[{i}]);\n'
s=s[:where]+printing+s[where:];(dst/'main.c').write_text(s)
