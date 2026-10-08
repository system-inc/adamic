interface Base { readonly kind: 'block' | 'comma' | 'case' | 'default' | 'identifier' | 'chain' | 'command' | 'label' | 'start'; }
interface Range { readonly pos: number; readonly end: number; }
interface NodeArray<T> extends ReadonlyArray<T>, Range { readonly hasTrailingComma: boolean; }
interface Expression extends Base { readonly pos: number; }
interface CaseClause extends Base { readonly kind: 'case'; readonly pos: number; readonly expression: Expression; }
interface DefaultClause extends Base { readonly kind: 'default'; readonly pos: number; }
type CaseOrDefaultClause = CaseClause | DefaultClause;
interface CaseBlock extends Base { readonly kind: 'block'; readonly clauses: NodeArray<CaseOrDefaultClause>; }
interface CommaListExpression extends Base { readonly kind: 'comma'; readonly elements: NodeArray<Expression>; }
interface DiagnosticMessageChain extends Base { readonly kind: 'chain'; messageText: string; category: number; code: number; next?: DiagnosticMessageChain[]; }
interface ParsedCommandLine extends Base { readonly kind: 'command'; fileNames: string[]; }
interface FlowNodeBase extends Base { flags: number; id: number; node: unknown; antecedent: FlowNode | FlowNode[] | undefined; }
interface FlowStart extends FlowNodeBase { readonly kind: 'start'; node: undefined; antecedent: undefined; }
interface FlowLabel extends FlowNodeBase { readonly kind: 'label'; node: undefined; antecedent: FlowNode[] | undefined; }
type FlowNode = FlowStart | FlowLabel;
function caseBlock(node: Base): CaseBlock { return node as CaseBlock; }
function commaList(node: Base): CommaListExpression { return node as CommaListExpression; }
function chain(value: Base): DiagnosticMessageChain { return value as DiagnosticMessageChain; }
function commandLine(value: Base): ParsedCommandLine { return value as ParsedCommandLine; }
function label(value: Base): FlowLabel { return value as FlowLabel; }
const raw = {kind: 'block' as const, clauses: Object.assign([{kind: 'other' as const, pos: 3}], {pos: 0, end: 1, hasTrailingComma: false})};
const clauses = caseBlock(raw).clauses;
console.log(`${clauses[0]!.pos}`);
