# Generates the microbenchmarks: for R reads per visit, a checked-view loop and a proven-type loop.
import sys
R = int(sys.argv[1]); mode = sys.argv[2]   # mode: view | plain | cast  (cast = tag check, no field read)
fields = "abcdefgh"
reads = " + ".join(f"v.{fields[i % 8]}" for i in range(R)) if R else "0"
print(f"""interface Node {{ readonly kind: number; }}
interface Ident extends Node {{ readonly kind: 1; readonly a: number; readonly b: number; readonly c: number; readonly d: number; readonly e: number; readonly f: number; readonly g: number; readonly h: number; }}
interface Other extends Node {{ readonly kind: 2; }}
const size = 1024;
const count = 40000;
""")
if mode == "view":
    print("""const nodes: Node[] = [];
for (let i = 0; i < size; i++) {
    const made: Ident = { kind: 1, a: i, b: 1, c: 2, d: 3, e: 4, f: 5, g: 6, h: 7 };
    nodes.push(made);
}
let total = 0;
for (let round = 0; round < count; round++) {
    for (let i = 0; i < size; i++) {
        const n = nodes[i];
        if (n === undefined) { continue; }
        const v = n as Ident;
        total += %s;
    }
}
console.log(String(total));""" % reads)
else:
    print("""const nodes: Ident[] = [];
for (let i = 0; i < size; i++) {
    nodes.push({ kind: 1, a: i, b: 1, c: 2, d: 3, e: 4, f: 5, g: 6, h: 7 });
}
let total = 0;
for (let round = 0; round < count; round++) {
    for (let i = 0; i < size; i++) {
        const v = nodes[i];
        if (v === undefined) { continue; }
        total += %s;
    }
}
console.log(String(total));""" % reads)
