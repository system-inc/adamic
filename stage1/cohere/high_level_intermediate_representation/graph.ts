// Adapter only: all graph algorithms are imported from the existing SSA port.
import { panic } from 'adamic';
import { construct } from '../static_single_assignment/construct.ts';
import { reversePostorder, markPredecessors, markEvaluationOrder } from '../static_single_assignment/graph.ts';
import type { GraphInterface, EdgeType } from '../static_single_assignment/static_single_assignment.ts';
import { BlockIndex, IdentifierIndex, DeclarationIndex } from '../arena/arena_index.a';
import { HIRFunction, BasicBlock, HIRArena, testPlace, testBlock } from './core.ts';
import type { PlaceInterface, FunctionIndex, BlockIndex as HIRBlockIndex }  from './core.ts';
function graphFor(fnOwner: HIRFunction): GraphInterface<HIRFunction, BasicBlock, PlaceInterface> { return {
    entry: (fn) => BlockIndex.read(fn.blockIndices, fn.entry) + 1,
    blockBound: (fn) => fn.nextBlock,
    block: (fn, id) => fn.retained.has(id) ? fn.block(fn.blockAt(id)) : undefined,
    blocks: (fn) => fn.blockOrder.map((id) => fn.block(id)),
    setBlocks: function(fn, blocks) { fn.blockOrder = blocks.map((block) => block.id); },
    retain: function(fn, keep) { for(const id of fn.retained) { if(!keep(id)) { fn.retained.delete(id); } } },
    placeholder: function(fn, block) {
        const placeholder = new BasicBlock(fn.blockIndices, block.id, { kind: 'Unreachable' }, block.kind);
        placeholder.predecessors = [...block.predecessors];
        fn.blockTable[BlockIndex.read(fn.blockIndices, block.id)] = placeholder;
        return placeholder;
    },
    id: (block) => BlockIndex.read(block.arenaIndices, block.id) + 1,
    predecessors: (block) => block.predecessors.map((id) => BlockIndex.read(block.arenaIndices, id) + 1),
    setPredecessors: function(block: BasicBlock, predecessors: readonly number[]) {
        const indices: BlockIndex[] = [];
        for(const id of predecessors) { const index = block.arenaIndices[id - 1] ?? panic('missing predecessor index'); BlockIndex.read(block.arenaIndices, index); indices.push(index); }
        block.predecessors = indices;
    },
    phis: (block) => block.phis,
    setPhis: function(block, phis) { block.phis = phis; },
    eachEdge: function(block, visit) {
        const terminal = block.terminal;
        const edge = (index: HIRBlockIndex, kind: EdgeType): void => { visit(BlockIndex.read(block.arenaIndices, index) + 1, kind); };
        if(terminal.kind === 'If' || terminal.kind === 'Branch' || terminal.kind === 'Optional' || terminal.kind === 'Logical' || terminal.kind === 'Ternary' || terminal.kind === 'While' || terminal.kind === 'DoWhile' || terminal.kind === 'For' || terminal.kind === 'ForOf' || terminal.kind === 'ForIn' || terminal.kind === 'Switch' || terminal.kind === 'Label' || terminal.kind === 'Try') { edge(terminal.fallthrough ?? panic('missing fallthrough'), 'Fallthrough'); }
        if(terminal.kind === 'Goto') { edge(terminal.block ?? panic('missing goto target'), 'Real'); }
        else if(terminal.kind === 'If' || terminal.kind === 'Branch') { edge(terminal.consequent ?? panic('missing consequent'), 'Real'); edge(terminal.alternate ?? panic('missing alternate'), 'Real'); }
        else if(terminal.kind === 'DoWhile') { edge(terminal.loop ?? panic('missing do loop'), 'Real'); }
        else if(terminal.kind === 'For' || terminal.kind === 'ForOf' || terminal.kind === 'ForIn') { edge(terminal.init ?? panic('missing loop init'), 'Real'); }
        else if(terminal.kind === 'Label') { edge(terminal.block ?? panic('missing label block'), 'Real'); }
        else if(terminal.kind === 'Try') { edge(terminal.block ?? panic('missing try block'), 'Real'); edge(terminal.handler ?? panic('missing handler'), 'Real'); }
        else if(terminal.kind === 'Switch') { const cases = terminal.cases ?? panic('missing cases'); for(const clause of cases) { edge(clause.block, 'Real'); } if(!cases.some((clause) => clause.test === undefined)) { edge(terminal.fallthrough ?? panic('missing switch exit'), 'Real'); } }
        else if(terminal.kind === 'Optional' || terminal.kind === 'Logical' || terminal.kind === 'Ternary' || terminal.kind === 'While') { edge(testBlock(terminal), 'Real'); }
    },
    endsInReturn: (block) => block.terminal.kind === 'Return',
    instructionCount: (_fn, block) => block.instructions.length,
    eachInstructionPlace: function(fn, block, index, visit) {
        const instruction = fn.instruction(block.instructions[index] ?? panic('missing instruction id'));
        instruction.lvalue = visit(instruction.lvalue, 'Define');
        if((instruction.value.kind === 'LoadLocal' || instruction.value.kind === 'LoadContext')) { instruction.value.place = visit(instruction.value.place, 'Use'); }
        if(instruction.value.kind === 'UnaryExpression') { instruction.value.value = visit(instruction.value.value, 'Use'); }
        if(instruction.value.kind === 'BinaryExpression') {
            instruction.value.left = visit(instruction.value.left, 'Use');
            instruction.value.right = visit(instruction.value.right, 'Use');
        }
        const value = instruction.value;
        if(value.kind === 'TypeCastExpression' || value.kind === 'Await') { value.value = visit(value.value, 'Use'); }
        if(value.kind === 'PropertyDelete' || value.kind === 'ComputedDelete') { value.object = visit(value.object, 'Use'); }
        if(value.kind === 'ComputedDelete') { value.property = visit(value.property, 'Use'); }
        if(value.kind === 'TaggedTemplateExpression') { value.tag = visit(value.tag, 'Use'); }
        if(value.kind === 'TemplateLiteral' || value.kind === 'TaggedTemplateExpression') { for(let index = 0; index < value.subexprs.length; index++) { value.subexprs[index] = visit(value.subexprs[index] ?? panic('missing template operand'), 'Use'); } }
        if(value.kind === 'StoreGlobal' || value.kind === 'StoreLocal' || value.kind === 'StoreContext' || value.kind === 'PrefixUpdate' || value.kind === 'PostfixUpdate' || value.kind === 'PropertyStore' || value.kind === 'ComputedStore') { value.value = visit(value.value, 'Use'); }
        if(value.kind === 'PropertyLoad' || value.kind === 'ComputedLoad' || value.kind === 'PropertyStore' || value.kind === 'ComputedStore') { value.object = visit(value.object, 'Use'); }
        if(value.kind === 'ComputedLoad' || value.kind === 'ComputedStore') { value.property = visit(value.property, 'Use'); }
        if(value.kind === 'CallExpression' || value.kind === 'NewExpression') { value.callee = visit(value.callee, 'Use'); }
        if(value.kind === 'MethodCall') { value.receiver = visit(value.receiver, 'Use'); value.property = visit(value.property, 'Use'); }
        if(value.kind === 'CallExpression' || value.kind === 'NewExpression' || value.kind === 'MethodCall') { for(const argument of value.args) { argument.place = visit(argument.place, 'Use'); } }
        if(value.kind === 'GetIterator' || value.kind === 'NextPropertyOf') { value.value = visit(value.value, 'Use'); }
        if(value.kind === 'IteratorNext') { value.iterator = visit(value.iterator, 'Use'); value.collection = visit(value.collection, 'Use'); }
        if(value.kind === 'ObjectExpression') { for(const property of value.properties) { if(property.computedKey !== undefined) { property.computedKey = visit(property.computedKey, 'Use'); } property.value = visit(property.value, 'Use'); } }
        if(value.kind === 'ArrayExpression') { for(const element of value.elements) { if(!element.hole) { element.place = visit(element.place, 'Use'); } } }
        if(value.kind === 'JsxExpression') { if(value.tag.place !== undefined) { value.tag.place = visit(value.tag.place, 'Use'); } for(const prop of value.props) { prop.value = visit(prop.value, 'Use'); } }
        if(value.kind === 'JsxExpression' || value.kind === 'JsxFragment') { for(let index = 0; index < value.children.length; index++) { value.children[index] = visit(value.children[index] ?? panic('missing jsx child'), 'Use'); } }
        if(value.kind === 'FunctionExpression') { for(let index = 0; index < value.captures.length; index++) { value.captures[index] = visit(value.captures[index] ?? panic('missing capture'), 'Use'); } }
        if(value.kind === 'DeclareLocal' || value.kind === 'StoreLocal' || value.kind === 'StoreContext' || value.kind === 'PrefixUpdate' || value.kind === 'PostfixUpdate') { value.lvalue = visit(value.lvalue, 'Define'); }

    },
    isContextStore: (fn, block, index) => (fn.instruction(block.instructions[index] ?? panic('missing instruction id'))).value.kind === 'StoreContext',
    contextStoreDefines: (fn, block, index, place) => place.identifier === (fn.instruction(block.instructions[index] ?? panic('missing instruction id'))).lvalue.identifier,
    setInstructionOrder: function(fn, block, index, order) {
        const instruction = fn.instruction(block.instructions[index] ?? panic('missing instruction id'));
        instruction.order = order;
    },
    eachTerminalPlace: function(block, visit) {
        const terminal = block.terminal;
        if(terminal.kind === 'Return' || terminal.kind === 'Throw') { terminal.value = visit(terminal.value ?? panic('missing terminal value'), 'Use'); }
        else if(terminal.kind === 'If' || terminal.kind === 'Branch' || terminal.kind === 'Switch') { terminal.testPlace = visit(testPlace(terminal), 'Use'); }
        if(terminal.kind === 'Switch') { for(const clause of terminal.cases ?? panic('missing cases')) { if(clause.test !== undefined) { clause.test = visit(clause.test, 'Use'); } } }
        if(terminal.kind === 'Try' && terminal.handlerBinding !== undefined) { terminal.handlerBinding = visit(terminal.handlerBinding, 'Define'); }
    },
        // Terminal case/handler operands are owned by their tagged arena record.
    setTerminalOrder: function(block, order) { block.terminalOrder = order; },
    params: (fn) => fn.params,
    returns: (fn) => fn.returns,
    setReturns: function(fn, place) { fn.returns = place; },
    declaration: (fn, id) => DeclarationIndex.read(fn.declarationIndices, fn.identifier(fn.identifierAt(id)).declaration) + 1,
    contextual: (fn, declaration) => fn.contextDeclarations.has(fn.declarationAt(declaration)),
    mint: (fn, original) => fn.mint(fn.identifierAt(original)).slot,
    named: (fn, id) => fn.identifier(fn.identifierAt(id)).name !== '',
    placeString: (_fn, id) => `$${id}`,
    identifierOf: (place) => IdentifierIndex.read(fnOwner.identifierIndices, place.identifier),
    withIdentifier: (place, identifier) => ({ identifier: fnOwner.identifierAt(identifier), effect: place.effect, reactive: place.reactive, start: place.start, end: place.end }),
}; }
export function constructHIR(arena: HIRArena, index: FunctionIndex): void {
    const fn = arena.read(index);
    const hirGraph = graphFor(fn);
    reversePostorder(hirGraph, fn);
    markPredecessors(hirGraph, fn);
    markEvaluationOrder(hirGraph, fn);
    construct(hirGraph, fn);
    for(const nested of fn.functions) { constructHIR(arena, nested); }
}
