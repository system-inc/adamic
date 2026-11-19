# The literal-run bench (#zpzm9ta): a read of v.kind through a checked view whose kind is an N-literal union, every
# value the last literal (the || chain's worst case), against a proven read of the same field. mode: view or plain;
# read: 1 reads v.kind, 0 adds a constant, so each variant is netted of its own loop and cast.
import sys
N, mode, read = int(sys.argv[1]), sys.argv[2], sys.argv[3] == "1"
kinds = " | ".join(str(k) for k in range(1, N + 1))
body = "v.kind" if read else "1"
print(f"""interface Node {{ readonly kind: number; }}
interface Wide extends Node {{ readonly kind: {kinds}; readonly a: number; }}
const size = 1024;
const count = 40000;
""")
if mode == "view":
    print(f"""const nodes: Node[] = [];
for (let i = 0; i < size; i++) {{
    const made: Wide = {{ kind: {N}, a: i }};
    nodes.push(made);
}}
let total = 0;
for (let round = 0; round < count; round++) {{
    for (let i = 0; i < size; i++) {{
        const n = nodes[i];
        if (n === undefined) {{ continue; }}
        const v = n as Wide;
        total += {body};
    }}
}}
console.log(String(total));""")
else:
    print(f"""const nodes: Wide[] = [];
for (let i = 0; i < size; i++) {{
    nodes.push({{ kind: {N}, a: i }});
}}
let total = 0;
for (let round = 0; round < count; round++) {{
    for (let i = 0; i < size; i++) {{
        const v = nodes[i];
        if (v === undefined) {{ continue; }}
        total += {body};
    }}
}}
console.log(String(total));""")
