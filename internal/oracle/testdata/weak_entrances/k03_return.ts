type Mode = "A" | "B";
interface Weak { overrideImportMode?: Mode; }
type Rec = { [key: string]: string };
function make(): Rec { const made: Rec = {}; made["overrideImportMode"] = "esm"; return made; }
function flag(): boolean { return true; }
const rec: Rec = make();
function give(): Weak { return rec; }
const w = give();
console.log(`${typeof w.overrideImportMode}:${String(w.overrideImportMode)}`);
