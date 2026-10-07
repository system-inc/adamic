/* eslint-disable max-classes-per-file -- The driver holds its IR's classes, as the single assignment slice's does. */
/*
 * The port's driver: it reads the cases file named by its one argument, builds each function it
 * describes on an IR of its own, runs the alias graph and the ranges on it, and prints what they made,
 * so its output can be held byte for byte to Go cohere's mutation_aliasing module running the same
 * passes on the same functions (mutation_aliasing_test.go, whose Go side reads and prints the same way).
 *
 * The IR is the Go module's test IR (mutation_aliasing_test.go) with what cohere's high-level IR answers
 * beside it: edges of every kind, an absent instruction, and a closure rule. A block holds phis and
 * instructions, each instruction its places in the order it visits them, its effects, the value it
 * stores into a captured binding if it is such a store, and what Closure answers at its Creates.
 *
 * A cases file is lines of fields separated by spaces, the first naming the record. A function is:
 *
 * 	function <name>                    begins a function
 * 	entry <block>                      the block control begins at
 * 	bound <n>                          one past the largest block id
 * 	frozen-parameters                  its parameters arrive Frozen (GraphInterface.parametersFrozen)
 * 	parameters-defined-on-entry        OptionsInterface.parametersDefinedOnEntry
 * 	context-kind <identifier> <kind>   an entry of OptionsInterface.contextKinds
 * 	param <identifier>                 the next parameter
 * 	context <identifier>               the next captured value (GraphInterface.context)
 * 	returns <identifier>               the place every return stores into
 * 	block <id> <terminal order>        the next block, in the function's order; the records after it are
 * 	                                   its own
 * 	edge <block> <kind>                its next edge: Real, Fallthrough or Exceptional (read by finalize)
 * 	return <identifier>                its terminal returns that value
 * 	phi <identifier> <predecessor>=<identifier>...   a phi and its operands
 * 	instruction <order>                its next instruction, at that evaluation order; `-` for one the
 * 	                                   function does not hold, whose order the IR doesn't answer
 * 	use <identifier>, define <identifier>   a place the instruction visits, in order
 * 	effect <kind> <into> <from or -> <value kind>   its next effect, kinds named as Go's String names them
 * 	stored-context <identifier>        the value it stores into a captured binding
 * 	closure <into> <rule>              what closure answers at a Create of into: `answer 0|1`, the frozen
 * 	                                   answer recorded from cohere's own IR; `immutable`, frozen when every
 * 	                                   capture's kind is immutable; or `nested <function>`, frozen when
 * 	                                   every capture's kind is immutable and that function's alias graph,
 * 	                                   built with those kinds as its context's, collects no mutation
 * 	capture <identifier>               the closure's next capture
 * 	expect <identifier> <start> <end>, expect-length <n>   what cohere's own pass gave the function,
 * 	                                   which Go's side checks it reproduces; the port reads past them
 * 	passes <ranges|finalize-ranges>    ends the function: run the pass on it as it is, or finalize it first
 *
 * Every function is read before any runs, so `nested` can name one later in the file. Three records
 * stand alone: `vocabulary` prints the effect kinds' and value kinds' names and the range gaps, `probe
 * range <start> <end> <order>` a range's predicates, and `probe phi-opens-before|phi-opened-range
 * <start> <end> <first order>` the phi helpers.
 *
 * The output, for each function, is `function <name>`, `order` with the block ids in their order, then
 * per block `block <id> first=<its first order> terminal=<order> instructions=<orders>`; the alias
 * graph's nodes by id, `node <id> <object|phi> created-from=... captures=... aliases=... maybe-aliases=...
 * edges=...`, an entry being `<identifier>@<index>` and an edge `<identifier>@<index>:<kind>`; each
 * collected `mutation <index> <identifier> <end> <transitive> <kind>` and `mutations <count>`; then per
 * value the function names, in id order, `value <id> range=<start>,<end> widened=<start>,<end>
 * kind=<kind> contains=<bits>`, widened being the table after the widening alone and the bits contains
 * at start - 1, start, end - 1 and end; then `length <n>`, `invalid <identifiers>` and `idempotent <0|1>`.
 * `-` is an unset range or an empty list.
 *
 * 	node oracle/node.mjs stage1/cohere/mutation_aliasing/main.ts stage1/cohere/mutation_aliasing/sample-cases.txt
 */

import { panic, programArguments, readTextFile } from 'adamic';
import { markEvaluationOrder, markPredecessors, reversePostorder } from '../static_single_assignment/graph.ts';
import { withPhiOperand } from '../static_single_assignment/phi.ts';
import type {
    BlockIdType,
    EdgeType,
    EvaluationOrderType,
    IdentifierIdType,
    PhiInterface,
    PhiOperandInterface,
    PlaceVisitorType,
    RoleType,
} from '../static_single_assignment/static_single_assignment.ts';
import {
    aliasingEffectKindName,
    AliasingEffectKinds,
    effectValueKindName,
    EffectValueKinds,
    isAliasing,
    isMutation,
} from './mutation_aliasing.ts';
import type {
    AliasingEffectInterface,
    AliasingEffectKindType,
    ClosureInterface,
    EffectValueKindType,
    GraphInterface,
} from './mutation_aliasing.ts';
import {
    blockFirstOrder,
    buildAliasingGraph,
    inferMutableRanges,
    MutableRange,
    MutableRanges,
    phiOpenedRange,
    phiOpensBefore,
    rangeGaps,
    validateMutableRanges,
} from './ranges.ts';
import type { AliasingNode, RangeGapType } from './ranges.ts';

// OraclePlaceInterface names its value identifier rather than id, so no block (whose id is a number too)
// can be seen as a place, which stage 0's cycle rule would refuse a visit's place field for.
interface OraclePlaceInterface {
    readonly identifier: IdentifierIdType;
}

interface OracleVisitInterface {
    place: OraclePlaceInterface;
    readonly role: RoleType;
}

interface OracleClosureInterface {
    readonly into: IdentifierIdType;
    readonly rule: string;
    readonly answer: boolean;
    readonly nested: string;
    readonly captures: OraclePlaceInterface[];
}

class OracleInstruction {
    readonly visits: OracleVisitInterface[] = [];
    readonly effects: AliasingEffectInterface<OraclePlaceInterface>[] = [];
    storedContext: IdentifierIdType | undefined = undefined;
    readonly closures: OracleClosureInterface[] = [];
    // absent is an instruction the function doesn't hold, whose order the IR doesn't answer.
    readonly absent: boolean;
    order: EvaluationOrderType;

    constructor(order: EvaluationOrderType, absent: boolean) {
        this.order = order;
        this.absent = absent;
    }
}

interface OracleEdgeInterface {
    readonly to: BlockIdType;
    readonly kind: EdgeType;
}

class OracleBlock {
    readonly id: BlockIdType;
    readonly instructions: OracleInstruction[] = [];
    readonly edges: OracleEdgeInterface[] = [];
    returnValue: OraclePlaceInterface | undefined = undefined;
    predecessors: readonly BlockIdType[] = [];
    phis: readonly PhiInterface<OraclePlaceInterface>[] = [];
    terminalOrder: EvaluationOrderType;

    constructor(id: BlockIdType, terminalOrder: EvaluationOrderType) {
        this.id = id;
        this.terminalOrder = terminalOrder;
    }
}

class OracleFunction {
    readonly name: string;
    entry: BlockIdType = 0;
    bound = 0;
    blocks: readonly OracleBlock[] = [];
    readonly table = new Map<BlockIdType, OracleBlock>();
    readonly params: OraclePlaceInterface[] = [];
    readonly context: OraclePlaceInterface[] = [];
    returns: OraclePlaceInterface | undefined = undefined;
    frozenParameters = false;
    parametersDefinedOnEntry = false;
    readonly contextKinds = new Map<IdentifierIdType, EffectValueKindType>();
    passes = '';

    constructor(name: string) {
        this.name = name;
    }
}

// functions are every function the cases file holds, by name, for a closure's `nested` rule.
const functions = new Map<string, OracleFunction>();

function instructionOf(block: OracleBlock, index: number): OracleInstruction {
    return block.instructions[index] ?? panic(`no instruction ${index} in bb${block.id}`);
}

// visitAll visits every place in visits, in order, and writes back what each visit returns.
function visitAll(visits: OracleVisitInterface[], visit: PlaceVisitorType<OraclePlaceInterface>): void {
    for(const each of visits) {
        each.place = visit(each.place, each.role);
    }
}

// placeholderOf is the empty block the adapter's placeholder puts in place of one. A named function, since
// stage 0 refuses an arrow with a block body returning a class (the single assignment slice's gap 1).
function placeholderOf(fn: OracleFunction, block: OracleBlock): OracleBlock {
    const placeholder = new OracleBlock(block.id, 0);
    placeholder.predecessors = [...block.predecessors];
    // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- placeholder: the adapter puts the placeholder in the function's table, as GraphInterface asks.
    fn.table.set(block.id, placeholder);
    return placeholder;
}

// closureOf is what the adapter's closure answers: the instruction's first closure record for into,
// under its rule, building a nested function's alias graph through graph.
function closureOf(
    graph: GraphInterface<OracleFunction, OracleBlock, OraclePlaceInterface>,
    instruction: OracleInstruction,
    into: IdentifierIdType,
    kinds: ReadonlyMap<IdentifierIdType, EffectValueKindType>,
): ClosureInterface<OraclePlaceInterface> | undefined {
    for(const closure of instruction.closures) {
        if(closure.into !== into) {
            continue;
        }
        if(closure.rule === 'answer') {
            return { captures: closure.captures, frozen: closure.answer };
        }
        let immutable = closure.captures.length > 0;
        for(const capture of closure.captures) {
            if(!kinds.has(capture.identifier)) {
                immutable = false;
            }
        }
        if(closure.rule === 'immutable' || !immutable) {
            return { captures: closure.captures, frozen: immutable };
        }
        const nested = functions.get(closure.nested) ?? panic(`no function ${closure.nested}`);
        if(nested.context.length !== closure.captures.length) {
            return { captures: closure.captures, frozen: false };
        }
        const contextKinds = new Map<IdentifierIdType, EffectValueKindType>();
        for(let position = 0; position < nested.context.length; position++) {
            const context = nested.context[position] ?? panic('a context past its end');
            const capture = closure.captures[position] ?? panic('a capture past its end');
            contextKinds.set(context.identifier, kinds.get(capture.identifier) ?? 'Mutable');
        }
        const nestedGraph = buildAliasingGraph(graph, nested, { parametersDefinedOnEntry: false, contextKinds });
        return { captures: closure.captures, frozen: nestedGraph.mutationCount() === 0 };
    }
    return undefined;
}

const oracleGraph: GraphInterface<OracleFunction, OracleBlock, OraclePlaceInterface> = {
    entry: (fn) => fn.entry,
    blockBound: (fn) => fn.bound,
    block: (fn, id) => fn.table.get(id),
    blocks: (fn) => fn.blocks,
    setBlocks: function(fn, blocks) {
        // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- setBlocks writes the function's block array, as GraphInterface asks.
        fn.blocks = blocks;
    },
    retain: function(fn, keep) {
        for(const id of [...fn.table.keys()]) {
            if(!keep(id)) {
                // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- retain forgets the blocks keep declines, as GraphInterface asks.
                fn.table.delete(id);
            }
        }
    },
    placeholder: (fn, block) => placeholderOf(fn, block),
    id: (block) => block.id,
    predecessors: (block) => block.predecessors,
    setPredecessors: function(block, predecessors) {
        // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- setPredecessors writes the block's predecessors, as GraphInterface asks.
        block.predecessors = predecessors;
    },
    phis: (block) => block.phis,
    setPhis: function(block, phis) {
        // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- setPhis writes the block's phis, as GraphInterface asks.
        block.phis = phis;
    },
    eachEdge: function(block, visit) {
        for(const edge of block.edges) {
            visit(edge.to, edge.kind);
        }
    },
    endsInReturn: (block) => block.returnValue !== undefined,
    instructionCount: (_fn, block) => block.instructions.length,
    eachInstructionPlace: function(_fn, block, index, visit) {
        visitAll(instructionOf(block, index).visits, visit);
    },
    isContextStore: (_fn, block, index) => instructionOf(block, index).storedContext !== undefined,
    contextStoreDefines: () => false,
    setInstructionOrder: function(_fn, block, index, order) {
        instructionOf(block, index).order = order;
    },
    eachTerminalPlace: function() {
        // The oracle IR's terminals hold no places.
    },
    setTerminalOrder: function(block, order) {
        // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- setTerminalOrder writes the block's terminal order, as GraphInterface asks.
        block.terminalOrder = order;
    },
    params: (fn) => fn.params,
    returns: (fn) => fn.returns,
    setReturns: function(fn, place) {
        // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- setReturns writes the function's returns place, as GraphInterface asks.
        fn.returns = place;
    },
    declaration: (_fn, id) => id,
    contextual: () => false,
    mint: (fn, original) => panic(`${fn.name}: the ranges never mint a value (asked for ${original})`),
    named: () => true,
    placeString: (_fn, id) => `${id}`,
    identifierOf: (place) => place.identifier,
    withIdentifier: (_place, identifier) => ({ identifier }),

    instructionOrder: function(_fn, block, index) {
        const instruction = instructionOf(block, index);
        return instruction.absent ? undefined : instruction.order;
    },
    terminalOrder: (block) => block.terminalOrder,
    effects: (_fn, block, index) => instructionOf(block, index).effects,
    parametersFrozen: (fn) => fn.frozenParameters,
    context: (fn) => fn.context,
    returnValue: (block) => block.returnValue,
    storedContextValue: (_fn, block, index) => instructionOf(block, index).storedContext,
    closure: (_fn, block, index, into, kinds) => closureOf(oracleGraph, instructionOf(block, index), into, kinds),
};

function numberField(fields: readonly string[], index: number, line: string): number {
    const text = fields[index] ?? panic(`a record without field ${index}: ${line}`);
    const value = Number.parseInt(text, 10);
    if(Number.isNaN(value)) {
        panic(`field ${index} is no number: ${line}`);
    }
    return value;
}

function textField(fields: readonly string[], index: number, line: string): string {
    return fields[index] ?? panic(`a record without field ${index}: ${line}`);
}

function placeField(fields: readonly string[], index: number, line: string): OraclePlaceInterface {
    return { identifier: numberField(fields, index, line) };
}

function edgeKind(text: string, line: string): EdgeType {
    switch(text) {
        case 'Real':
        case 'Fallthrough':
        case 'Exceptional':
            return text;
    }
    return panic(`an unknown edge kind: ${line}`);
}

function effectKindNamed(text: string, line: string): AliasingEffectKindType {
    for(const kind of AliasingEffectKinds) {
        if(aliasingEffectKindName(kind) === text) {
            return kind;
        }
    }
    return panic(`an unknown effect kind: ${line}`);
}

function valueKindNamed(text: string, line: string): EffectValueKindType {
    for(const kind of EffectValueKinds) {
        if(effectValueKindName(kind) === text) {
            return kind;
        }
    }
    return panic(`an unknown value kind: ${line}`);
}

function bit(value: boolean): string {
    return value ? '1' : '0';
}

function rangeText(range: MutableRange): string {
    return range.isSet() ? `${range.start},${range.end}` : '-';
}

function listText(items: readonly string[]): string {
    return items.length === 0 ? '-' : items.join(',');
}

function entriesText(entries: ReadonlyMap<IdentifierIdType, number>): string {
    const items: string[] = [];
    for(const [id, index] of entries) {
        items.push(`${id}@${index}`);
    }
    return listText(items);
}

function nodeText(node: AliasingNode): string {
    const edges: string[] = [];
    for(const edge of node.edges) {
        edges.push(
            `${edge.node}@${edge.index}:${edge.kind === 'MaybeAlias' ? 'maybe-alias' : edge.kind === 'Alias' ? 'alias' : 'capture'}`,
        );
    }
    return `node ${node.id} ${node.value === 'Phi' ? 'phi' : 'object'} created-from=${entriesText(node.createdFrom)} captures=${entriesText(node.captures)} aliases=${entriesText(node.aliases)} maybe-aliases=${entriesText(node.maybeAliases)} edges=${listText(edges)}`;
}

// valuesOf are the identifiers a function names anywhere, ascending.
function valuesOf(fn: OracleFunction): IdentifierIdType[] {
    const seen = new Set<IdentifierIdType>();
    const add = function(id: IdentifierIdType): void {
        seen.add(id);
    };
    for(const place of fn.params) {
        add(place.identifier);
    }
    for(const place of fn.context) {
        add(place.identifier);
    }
    if(fn.returns !== undefined) {
        add(fn.returns.identifier);
    }
    for(const id of fn.contextKinds.keys()) {
        add(id);
    }
    for(const block of fn.blocks) {
        if(block.returnValue !== undefined) {
            add(block.returnValue.identifier);
        }
        for(const phi of block.phis) {
            add(phi.place.identifier);
            for(const operand of phi.operands) {
                add(operand.place.identifier);
            }
        }
        for(const instruction of block.instructions) {
            for(const visit of instruction.visits) {
                add(visit.place.identifier);
            }
            for(const effect of instruction.effects) {
                add(effect.into.identifier);
                if(effect.from !== undefined) {
                    add(effect.from.identifier);
                }
            }
            if(instruction.storedContext !== undefined) {
                add(instruction.storedContext);
            }
            for(const closure of instruction.closures) {
                for(const capture of closure.captures) {
                    add(capture.identifier);
                }
            }
        }
    }
    const values = [...seen];
    values.sort((left, right) => left - right);
    return values;
}

// describe runs the passes on a function and prints what they made.
function describe(fn: OracleFunction): void {
    if(fn.passes === 'finalize-ranges') {
        reversePostorder(oracleGraph, fn);
        markPredecessors(oracleGraph, fn);
        markEvaluationOrder(oracleGraph, fn);
    }
    else if(fn.passes !== 'ranges') {
        panic(`${fn.name}: unknown passes ${fn.passes}`);
    }
    const options = { parametersDefinedOnEntry: fn.parametersDefinedOnEntry, contextKinds: fn.contextKinds };

    console.log(`function ${fn.name}`);
    console.log(`order ${fn.blocks.map((block) => `${block.id}`).join(' ')}`);
    for(const block of fn.blocks) {
        const orders: string[] = [];
        for(const instruction of block.instructions) {
            orders.push(instruction.absent ? '-' : `${instruction.order}`);
        }
        console.log(
            `block ${block.id} first=${blockFirstOrder(oracleGraph, fn, block)} terminal=${block.terminalOrder} instructions=${listText(orders)}`,
        );
    }

    const graph = buildAliasingGraph(oracleGraph, fn, options);
    const nodeIds = [...graph.state.nodes.keys()];
    nodeIds.sort((left, right) => left - right);
    for(const id of nodeIds) {
        console.log(nodeText(graph.state.nodes.get(id) ?? panic(`no node ${id}`)));
    }
    for(const mutation of graph.mutations) {
        console.log(
            `mutation ${mutation.index} ${mutation.place} ${mutation.end} ${bit(mutation.transitive)} ${mutation.kind}`,
        );
    }
    console.log(`mutations ${graph.mutationCount()}`);

    const widened = new MutableRanges();
    graph.widen(widened);
    const ranges = inferMutableRanges(oracleGraph, fn, options);
    const again = inferMutableRanges(oracleGraph, fn, options);
    let idempotent = ranges.length() === again.length();
    for(const id of valuesOf(fn)) {
        const range = ranges.get(id);
        const other = again.get(id);
        if(range.start !== other.start || range.end !== other.end) {
            idempotent = false;
        }
        const contains = range.isSet()
            ? `${bit(range.contains(range.start - 1))}${bit(range.contains(range.start))}${bit(range.contains(range.end - 1))}${bit(range.contains(range.end))}`
            : '-';
        console.log(
            `value ${id} range=${rangeText(range)} widened=${rangeText(widened.get(id))} kind=${effectValueKindName(graph.kind(id))} contains=${contains}`,
        );
    }
    console.log(`length ${ranges.length()}`);
    console.log(`invalid ${listText(validateMutableRanges(ranges).map((id) => `${id}`))}`);
    console.log(`idempotent ${bit(idempotent)}`);
}

// rangeGapName is a gap's name as Go's side prints it. A switch, since stage 0 can't lower panic as an
// arm of a conditional (gap 3).
function rangeGapName(gap: RangeGapType): string {
    switch(gap) {
        case 'LoopCarriedInversion':
            return 'loop-carried-inversion';
    }
}

// standalone prints what a record outside any function asks for.
function standalone(fields: readonly string[], line: string): void {
    if(fields[0] === 'vocabulary') {
        console.log('vocabulary');
        for(const kind of AliasingEffectKinds) {
            console.log(
                `effect-kind ${aliasingEffectKindName(kind)} mutation=${bit(isMutation(kind))} aliasing=${bit(isAliasing(kind))}`,
            );
        }
        for(const kind of EffectValueKinds) {
            console.log(`value-kind ${effectValueKindName(kind)}`);
        }
        const gaps: string[] = [];
        for(const gap of rangeGaps()) {
            gaps.push(rangeGapName(gap));
        }
        console.log(`gaps ${listText(gaps)}`);
        return;
    }
    const what = textField(fields, 1, line);
    const range = new MutableRange(numberField(fields, 2, line), numberField(fields, 3, line));
    const at = numberField(fields, 4, line);
    switch(what) {
        case 'range':
            console.log(
                `range ${range.start} ${range.end} ${at} set=${bit(range.isSet())} valid=${bit(range.isValid())} contains=${bit(range.contains(at))}`,
            );
            return;
        case 'phi-opens-before':
            console.log(`phi-opens-before ${range.start} ${range.end} ${at} ${bit(phiOpensBefore(range, at))}`);
            return;
        case 'phi-opened-range': {
            const opened = phiOpenedRange(range, at);
            console.log(
                `phi-opened-range ${range.start} ${range.end} ${at} ${opened === undefined ? '-' : `${opened.start},${opened.end}`}`,
            );
            return;
        }
    }
    panic(`an unknown probe: ${line}`);
}

// ItemInterface is one thing the cases file asks for, in its order: a function to describe, or a record
// standing alone.
interface ItemInterface {
    readonly fn: OracleFunction | undefined;
    readonly fields: readonly string[];
    readonly line: string;
}

// readCases reads a cases file's records into what it asks for, in its order, and every function it
// holds into functions. A function rather than the module's own code, since the cycle rule takes what
// module-level code writes as written into what no function made (`GAPS.md`, the cycle rule's cost).
function readCases(text: string): ItemInterface[] {
    const items: ItemInterface[] = [];
    let current: OracleFunction | undefined;
    let currentBlock: OracleBlock | undefined;
    let currentInstruction: OracleInstruction | undefined;
    let currentClosure: OracleClosureInterface | undefined;
    for(const line of text.split('\n')) {
        if(line === '') {
            continue;
        }
        const fields = line.split(' ');
        const record = fields[0] ?? '';
        if(record === 'function') {
            const name = textField(fields, 1, line);
            if(functions.has(name)) {
                panic(`a second function named ${name}`);
            }
            current = new OracleFunction(name);
            functions.set(name, current);
            items.push({ fn: current, fields, line });
            currentBlock = undefined;
            currentInstruction = undefined;
            currentClosure = undefined;
            continue;
        }
        if(record === 'vocabulary' || record === 'probe') {
            if(current !== undefined) {
                panic(`a standalone record inside a function: ${line}`);
            }
            items.push({ fn: undefined, fields, line });
            continue;
        }
        const fn = current ?? panic(`a record before any function: ${line}`);
        switch(record) {
            case 'entry':
                fn.entry = numberField(fields, 1, line);
                break;
            case 'bound':
                fn.bound = numberField(fields, 1, line);
                break;
            case 'frozen-parameters':
                fn.frozenParameters = true;
                break;
            case 'parameters-defined-on-entry':
                fn.parametersDefinedOnEntry = true;
                break;
            case 'context-kind':
                fn.contextKinds.set(numberField(fields, 1, line), valueKindNamed(textField(fields, 2, line), line));
                break;
            case 'param':
                fn.params.push(placeField(fields, 1, line));
                break;
            case 'context':
                fn.context.push(placeField(fields, 1, line));
                break;
            case 'returns':
                fn.returns = placeField(fields, 1, line);
                break;
            case 'block': {
                const block = new OracleBlock(numberField(fields, 1, line), numberField(fields, 2, line));
                fn.blocks = [...fn.blocks, block];
                fn.table.set(block.id, block);
                currentBlock = block;
                currentInstruction = undefined;
                currentClosure = undefined;
                break;
            }
            case 'edge':
                (currentBlock ?? panic(`an edge outside a block: ${line}`)).edges.push({
                    to: numberField(fields, 1, line),
                    kind: edgeKind(textField(fields, 2, line), line),
                });
                break;
            case 'return':
                (currentBlock ?? panic(`a return outside a block: ${line}`)).returnValue = placeField(fields, 1, line);
                break;
            case 'phi': {
                const block = currentBlock ?? panic(`a phi outside a block: ${line}`);
                let operands: PhiOperandInterface<OraclePlaceInterface>[] = [];
                for(let index = 2; index < fields.length; index++) {
                    const operand = (fields[index] ?? '').split('=');
                    operands = withPhiOperand(operands, numberField(operand, 0, line), placeField(operand, 1, line));
                }
                block.phis = [...block.phis, { place: placeField(fields, 1, line), operands }];
                break;
            }
            case 'instruction': {
                const absent = textField(fields, 1, line) === '-';
                const instruction = new OracleInstruction(absent ? 0 : numberField(fields, 1, line), absent);
                (currentBlock ?? panic(`an instruction outside a block: ${line}`)).instructions.push(instruction);
                currentInstruction = instruction;
                currentClosure = undefined;
                break;
            }
            case 'use':
            case 'define':
                (currentInstruction ?? panic(`a place outside an instruction: ${line}`)).visits.push({
                    place: placeField(fields, 1, line),
                    role: record === 'use' ? 'Use' : 'Define',
                });
                break;
            case 'effect':
                (currentInstruction ?? panic(`an effect outside an instruction: ${line}`)).effects.push({
                    kind: effectKindNamed(textField(fields, 1, line), line),
                    into: placeField(fields, 2, line),
                    from: textField(fields, 3, line) === '-' ? undefined : placeField(fields, 3, line),
                    value: valueKindNamed(textField(fields, 4, line), line),
                });
                break;
            case 'stored-context':
                (currentInstruction ?? panic(`a stored context outside an instruction: ${line}`)).storedContext =
                    numberField(fields, 1, line);
                break;
            case 'closure': {
                const rule = textField(fields, 2, line);
                if(rule !== 'answer' && rule !== 'immutable' && rule !== 'nested') {
                    panic(`an unknown closure rule: ${line}`);
                }
                const closure: OracleClosureInterface = {
                    into: numberField(fields, 1, line),
                    rule,
                    answer: rule === 'answer' && numberField(fields, 3, line) === 1,
                    nested: rule === 'nested' ? textField(fields, 3, line) : '',
                    captures: [],
                };
                (currentInstruction ?? panic(`a closure outside an instruction: ${line}`)).closures.push(closure);
                currentClosure = closure;
                break;
            }
            case 'capture':
                (currentClosure ?? panic(`a capture outside a closure: ${line}`)).captures.push(
                    placeField(fields, 1, line),
                );
                break;
            case 'expect':
            case 'expect-length':
                // What cohere's own pass gave the function, which Go's side checks it reproduces.
                break;
            case 'passes':
                fn.passes = textField(fields, 1, line);
                current = undefined;
                break;
            default:
                panic(`an unknown record: ${line}`);
        }
    }
    return items;
}

const casesPath = programArguments()[0] ?? panic('usage: main.ts <cases file>');
const read = readTextFile(casesPath);
if(read.kind === 'Error') {
    panic(read.message);
}

const items = readCases(read.text);
for(const item of items) {
    if(item.fn === undefined) {
        standalone(item.fields, item.line);
    }
    else {
        describe(item.fn);
    }
}
