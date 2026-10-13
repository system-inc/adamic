interface FlowEnd { node: undefined }
interface FlowAssignment { node: { readonly kind: string } | undefined }
type FlowNode = FlowEnd | FlowAssignment;
const narrow: FlowEnd = { node: undefined };
function store(view: FlowNode): void { view.node = { kind: 'identifier' }; }
store(narrow);
console.log((narrow.node === undefined).toString());
