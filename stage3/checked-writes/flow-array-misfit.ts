interface FlowEnd { antecedents: undefined }
interface FlowLabel { antecedents: number[] | undefined }
type FlowNode = FlowEnd | FlowLabel;
const narrow: FlowEnd = { antecedents: undefined };
function store(view: FlowNode): void { view.antecedents = [1]; }
store(narrow);
console.log((narrow.antecedents === undefined).toString());
