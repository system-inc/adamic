type Mode = "A" | "B";
interface Weak { overrideImportMode?: Mode; }
type Rec = { [key: string]: string };
function make(): Rec { const made: Rec = {}; made["overrideImportMode"] = "esm"; return made; }
function flag(): boolean { return true; }
const rec: Rec = make();
const map = new Map<string, Weak>();
map.set("k", rec);
const w = map.get("k");
if (w !== undefined) { console.log(`${typeof w.overrideImportMode}:${String(w.overrideImportMode)}`); }
