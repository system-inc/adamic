interface Label { readonly flags: number; readonly node: undefined; }
interface Assignment { readonly node: string; readonly flags: number; }
let currentFlow: Label | Assignment = { flags: 1, node: undefined };
function bind(): void { currentFlow = { node: `assigned${2}`, flags: 12 }; }
if (currentFlow.node === undefined) {
    bind();
    console.log(`${currentFlow.flags}`);
}
// Keep numeric-enum combinations distinct from a member-specific literal tag.
enum FlowFlags { Start = 1, Assignment = 8, Referenced = 4 }
class FlowLabel { readonly flags: FlowFlags = FlowFlags.Start; }
let enumFlow: unknown = new FlowLabel();
function bindEnum(): void { enumFlow = { flags: FlowFlags.Assignment | FlowFlags.Referenced }; }
if (enumFlow instanceof FlowLabel) {
    bindEnum();
    console.log(`${enumFlow.flags}`);
}
