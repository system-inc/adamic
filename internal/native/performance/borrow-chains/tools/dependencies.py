"""Count the owned Parser.node returns left for the accessor-return unit."""
from pathlib import Path
import json,re,subprocess,sys
artifact=Path(sys.argv[1]);source=(artifact/'visited.c').read_text()
source=source.replace('static size_t chain_nodes;','static size_t chain_nodes, chain_node_calls;').replace('chain_nodes); }','chain_nodes); fprintf(stderr,"accessor_calls=%zu\\n",chain_node_calls); }')
source,count=re.subn(r'(static adamic_object \* adamic_function_\d+_Parser_node\([^\n]*\) \{)',r'\1\n chain_node_calls++;',source);assert count==1
path=Path('/tmp/borrow-chains-dependencies.c');path.write_text(source)
command=json.loads(Path('/tmp/borrow-chains-clang-commands.json').read_text())[-1]
command[command.index('-o')+1]='/tmp/borrow-chains-dependencies';command[command.index(str(artifact/'visited.c'))]=str(path)
with open('/tmp/borrow-chains-dependencies-build.log','wb') as log:subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,check=True)
with open('/tmp/borrow-chains-dependencies.stdout','wb') as output,open('/tmp/borrow-chains-dependencies.stderr','wb') as stderr:subprocess.run(['/tmp/borrow-chains-dependencies','--manifest',str(artifact/'compiler.txt'),'--count'],stdout=output,stderr=stderr,check=True)
print(Path('/tmp/borrow-chains-dependencies.stderr').read_text())
