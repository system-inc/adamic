type Mode = "A" | "B";
interface Weak { overrideImportMode?: Mode; }
type Rec = { [key: string]: string };
function make(): Rec { const made: Rec = {}; made["overrideImportMode"] = "esm"; return made; }
function flag(): boolean { return true; }
const rec: Rec = make();
function read(...ws: Weak[]): void { console.log(`${typeof ws[0].overrideImportMode}:${String(ws[0].overrideImportMode)}`); }
read(...[rec]);
