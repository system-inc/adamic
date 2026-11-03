type Mode = "A" | "B";
interface Weak { overrideImportMode?: Mode; }
type Rec = { [key: string]: string };
function make(): Rec { const made: Rec = {}; made["overrideImportMode"] = "esm"; return made; }
function flag(): boolean { return true; }
const rec: Rec = make();
class Holder { w: Weak = rec; }
const holder = new Holder();
console.log(`${typeof holder.w.overrideImportMode}:${String(holder.w.overrideImportMode)}`);
