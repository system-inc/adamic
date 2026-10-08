/*
 * The port's driver: it reads the cases file named by its one argument, builds each function it
 * describes on an IR of its own, runs the passes on it, and prints what they made, so its output can be
 * held byte for byte to Go cohere's module running the same passes on the same functions
 * (static_single_assignment_test.go, whose Go side reads and prints the same way).
 *
 * The IR is the Go module's test IR (static_single_assignment_test.go), with a terminal's places added,
 * which cohere's high-level IR has and the Go tests' IR leaves out: a block holds instructions, each
 * with the places it uses and defines, and a terminal with places of its own and the edges it names.
 *
 * A cases file is lines of fields separated by spaces, the first naming the record. A function is:
 *
 * 	function <name>                 begins a function
 * 	entry <block>                   the block control begins at
 * 	bound <n>                       one past the largest block id
 * 	identifier <declaration> <name> the next identifier, numbered from 0; `-` names a temporary
 * 	contextual <declaration>        a binding a nested function captures
 * 	param <identifier> <tag>        the next parameter
 * 	returns <identifier> <tag>      the place every return stores into
 * 	block <id>                      the next block, in creation order; the records after it are its own
 * 	return                          its terminal returns from the function
 * 	edge <block> <kind>             its next edge: Real, Fallthrough or Exceptional
 * 	instruction <result>            its next instruction, and the identifier a context store's own
 * 	                                result is (-1 for none)
 * 	context                         the instruction is a context store
 * 	use <identifier> <tag>          a place the instruction reads, in order
 * 	define <identifier> <tag>       a place the instruction writes, in order
 * 	terminal-use <identifier> <tag>, terminal-define <identifier> <tag>   the terminal's places
 * 	passes <finalize|construct>     ends the function: run the finalizer alone, or construction after it
 *
 * A tag stands in for what a real IR's place carries beside its value (cohere's effect, reactivity and
 * range), so the output shows it riding along through renaming; `-` is the empty tag.
 *
 * The output, for each function, is `function <name>`, then `order` with the block ids in their final
 * order, then each block: `block <id> <placeholder or -> predecessors=<ids> terminal=<order>
 * return=<0 or 1>`, its phis `phi <place> <predecessor>=<place>...`, its instructions `instruction
 * <order> <context or -> uses=<places> defines=<places>`, and `terminal uses=<places>
 * defines=<places>`. Then `params`, `returns`, `table` (the ids the function still finds a block for),
 * `identifiers <count>`, each `violation <kind> <identifier> <block> <detail>`, `stats <phis> <named
 * values> <uses>`, and `dominators <block>=<every block in order that dominates it>` per block. A place
 * prints as `<identifier>:<tag>`.
 *
 * 	node oracle/node.mjs stage1/cohere/static_single_assignment/main.ts stage1/cohere/static_single_assignment/sample-cases.txt
 */

import { panic, programArguments, readTextFile } from 'adamic';
import { construct } from './construct.ts';
import { markEvaluationOrder, markPredecessors, reversePostorder } from './graph.ts';
import type {
    BlockIdType,
    DeclarationIdType,
    EdgeType,
    EvaluationOrderType,
    GraphInterface,
    IdentifierIdType,
    PhiInterface,
    PlaceVisitorType,
} from './static_single_assignment.ts';
import { collectSingleAssignmentStats, computeDominance, verifySingleAssignment } from './verify.ts';

interface OraclePlaceInterface {
    readonly id: IdentifierIdType;
    readonly tag: string;
}

class OracleInstruction {
    readonly uses: OraclePlaceInterface[] = [];
    readonly defines: OraclePlaceInterface[] = [];
    contextStore = false;
    // result is the definition a context store's own result is, the one the verifier counts.
    readonly result: IdentifierIdType;
    order: EvaluationOrderType = 0;

    constructor(result: IdentifierIdType) {
        this.result = result;
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
    returns = false;
    predecessors: readonly BlockIdType[] = [];
    phis: readonly PhiInterface<OraclePlaceInterface>[] = [];
    readonly terminalUses: OraclePlaceInterface[] = [];
    readonly terminalDefines: OraclePlaceInterface[] = [];
    terminalOrder: EvaluationOrderType = 0;
    readonly placeholder: boolean;

    constructor(id: BlockIdType, placeholder: boolean) {
        this.id = id;
        this.placeholder = placeholder;
    }
}

interface OracleIdentifierInterface {
    readonly declaration: DeclarationIdType;
    readonly name: string;
}

class OracleFunction {
    readonly name: string;
    entry: BlockIdType = 0;
    bound = 0;
    blocks: readonly OracleBlock[] = [];
    readonly table = new Map<BlockIdType, OracleBlock>();
    readonly identifiers: OracleIdentifierInterface[] = [];
    readonly params: OraclePlaceInterface[] = [];
    returns: OraclePlaceInterface | undefined = undefined;
    readonly contextual = new Map<DeclarationIdType, boolean>();

    constructor(name: string) {
        this.name = name;
    }
}

function identifierOf(fn: OracleFunction, id: IdentifierIdType): OracleIdentifierInterface {
    return fn.identifiers[id] ?? panic(`no identifier ${id} in ${fn.name}`);
}

function instructionOf(block: OracleBlock, index: number): OracleInstruction {
    return block.instructions[index] ?? panic(`no instruction ${index} in bb${block.id}`);
}

// visitAll visits every place in places, in order, and writes back what each visit returns.
function visitAll(
    places: OraclePlaceInterface[],
    role: 'Use' | 'Define',
    visit: PlaceVisitorType<OraclePlaceInterface>,
): void {
    for(let index = 0; index < places.length; index++) {
        places[index] = visit(places[index] ?? panic('a place index past its end'), role);
    }
}

// placeholderOf is the empty unreachable block the adapter's placeholder puts in place of one.
function placeholderOf(fn: OracleFunction, block: OracleBlock): OracleBlock {
    const placeholder = new OracleBlock(block.id, true);
    placeholder.predecessors = [...block.predecessors];
    fn.table.set(block.id, placeholder);
    return placeholder;
}

const oracleGraph: GraphInterface<OracleFunction, OracleBlock, OraclePlaceInterface> = {
    entry: (fn) => fn.entry,
    blockBound: (fn) => fn.bound,
    block: (fn, id) => fn.table.get(id),
    blocks: (fn) => fn.blocks,
    setBlocks: function(fn, blocks) {
        fn.blocks = blocks;
    },
    retain: function(fn, keep) {
        for(const id of [...fn.table.keys()]) {
            if(!keep(id)) {
                fn.table.delete(id);
            }
        }
    },
    // A named function, since stage 0 refuses an arrow with a block body returning a class (gap 1).
    placeholder: (fn, block) => placeholderOf(fn, block),
    id: (block) => block.id,
    predecessors: (block) => block.predecessors,
    setPredecessors: function(block, predecessors) {
        block.predecessors = predecessors;
    },
    phis: (block) => block.phis,
    setPhis: function(block, phis) {
        block.phis = phis;
    },
    eachEdge: function(block, visit) {
        for(const edge of block.edges) {
            visit(edge.to, edge.kind);
        }
    },
    endsInReturn: (block) => block.returns,
    instructionCount: (_fn, block) => block.instructions.length,
    eachInstructionPlace: function(_fn, block, index, visit) {
        const instruction = instructionOf(block, index);
        visitAll(instruction.uses, 'Use', visit);
        visitAll(instruction.defines, 'Define', visit);
    },
    isContextStore: (_fn, block, index) => instructionOf(block, index).contextStore,
    contextStoreDefines: (_fn, block, index, place) => place.id === instructionOf(block, index).result,
    setInstructionOrder: function(_fn, block, index, order) {
        instructionOf(block, index).order = order;
    },
    eachTerminalPlace: function(block, visit) {
        visitAll(block.terminalUses, 'Use', visit);
        visitAll(block.terminalDefines, 'Define', visit);
    },
    setTerminalOrder: function(block, order) {
        block.terminalOrder = order;
    },
    params: (fn) => fn.params,
    returns: (fn) => fn.returns,
    setReturns: function(fn, place) {
        fn.returns = place;
    },
    declaration: (fn, id) => identifierOf(fn, id).declaration,
    contextual: (fn, declaration) => fn.contextual.get(declaration) === true,
    mint: function(fn, original) {
        const id = fn.identifiers.length;
        fn.identifiers.push(identifierOf(fn, original));
        return id;
    },
    named: (fn, id) => identifierOf(fn, id).name !== '',
    placeString: (fn, id) => `${identifierOf(fn, id).name}$${id}`,
    identifierOf: (place) => place.id,
    withIdentifier: (place, id) => ({ id, tag: place.tag }),
};

// field is a record's field at index, as a number or as a string.
function numberField(fields: readonly string[], index: number, line: string): number {
    const text = fields[index] ?? panic(`a record without field ${index}: ${line}`);
    const value = Number.parseInt(text, 10);
    if(Number.isNaN(value)) {
        panic(`field ${index} is no number: ${line}`);
    }
    return value;
}

function textField(fields: readonly string[], index: number, line: string): string {
    const text = fields[index] ?? panic(`a record without field ${index}: ${line}`);
    return text === '-' ? '' : text;
}

function placeField(fields: readonly string[], line: string): OraclePlaceInterface {
    return { id: numberField(fields, 1, line), tag: textField(fields, 2, line) };
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

function placeText(place: OraclePlaceInterface): string {
    return `${place.id}:${place.tag === '' ? '-' : place.tag}`;
}

function placesText(places: readonly OraclePlaceInterface[]): string {
    return places.map((place) => placeText(place)).join(',');
}

// describe prints what the passes made of a function.
function describe(fn: OracleFunction): void {
    console.log(`function ${fn.name}`);
    console.log(`order ${fn.blocks.map((block) => `${block.id}`).join(' ')}`);
    for(const block of fn.blocks) {
        console.log(
            `block ${block.id} ${block.placeholder ? 'placeholder' : '-'} predecessors=${block.predecessors.map((id) => `${id}`).join(',')} terminal=${block.terminalOrder} return=${block.returns ? '1' : '0'}`,
        );
        for(const phi of block.phis) {
            const operands = phi.operands
                .map((operand) => ` ${operand.predecessor}=${placeText(operand.place)}`)
                .join('');
            console.log(`phi ${placeText(phi.place)}${operands}`);
        }
        for(const instruction of block.instructions) {
            console.log(
                `instruction ${instruction.order} ${instruction.contextStore ? 'context' : '-'} uses=${placesText(instruction.uses)} defines=${placesText(instruction.defines)}`,
            );
        }
        console.log(`terminal uses=${placesText(block.terminalUses)} defines=${placesText(block.terminalDefines)}`);
    }
    console.log(`params ${placesText(fn.params)}`);
    console.log(`returns ${fn.returns === undefined ? '-' : placeText(fn.returns)}`);
    const found: string[] = [];
    for(let id = 0; id < fn.bound; id++) {
        if(fn.table.has(id)) {
            found.push(`${id}`);
        }
    }
    console.log(`table ${found.join(' ')}`);
    console.log(`identifiers ${fn.identifiers.length}`);
    for(const violation of verifySingleAssignment(oracleGraph, fn)) {
        console.log(`violation ${violation.kind} ${violation.identifier} ${violation.block} ${violation.detail}`);
    }
    const stats = collectSingleAssignmentStats(oracleGraph, fn);
    console.log(`stats ${stats.phis} ${stats.namedValues} ${stats.uses}`);
    const dominance = computeDominance(oracleGraph, fn);
    for(const block of fn.blocks) {
        const dominators = fn.blocks
            .filter((candidate) => dominance.dominates(candidate.id, block.id))
            .map((candidate) => `${candidate.id}`);
        console.log(`dominators ${block.id}=${dominators.join(',')}`);
    }
}

// run runs the passes the function's last record names, then prints what they made.
function run(fn: OracleFunction, passes: string, line: string): void {
    reversePostorder(oracleGraph, fn);
    markPredecessors(oracleGraph, fn);
    markEvaluationOrder(oracleGraph, fn);
    if(passes === 'construct') {
        construct(oracleGraph, fn);
    }
    else if(passes !== 'finalize') {
        panic(`unknown passes: ${line}`);
    }
    describe(fn);
}

const casesPath = programArguments()[0] ?? panic('usage: main.ts <cases file>');
const read = readTextFile(casesPath);
if(read.kind === 'Error') {
    panic(read.message);
}

let current: OracleFunction | undefined;
let currentBlock: OracleBlock | undefined;
let currentInstruction: OracleInstruction | undefined;
for(const line of read.text.split('\n')) {
    if(line === '') {
        continue;
    }
    const fields = line.split(' ');
    const record = fields[0] ?? '';
    if(record === 'function') {
        current = new OracleFunction(textField(fields, 1, line));
        currentBlock = undefined;
        currentInstruction = undefined;
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
        case 'identifier':
            fn.identifiers.push({ declaration: numberField(fields, 1, line), name: textField(fields, 2, line) });
            break;
        case 'contextual':
            fn.contextual.set(numberField(fields, 1, line), true);
            break;
        case 'param':
            fn.params.push(placeField(fields, line));
            break;
        case 'returns':
            fn.returns = placeField(fields, line);
            break;
        case 'block': {
            const block = new OracleBlock(numberField(fields, 1, line), false);
            fn.blocks = [...fn.blocks, block];
            fn.table.set(block.id, block);
            currentBlock = block;
            currentInstruction = undefined;
            break;
        }
        case 'return':
            (currentBlock ?? panic(`a return outside a block: ${line}`)).returns = true;
            break;
        case 'edge':
            (currentBlock ?? panic(`an edge outside a block: ${line}`)).edges.push({
                to: numberField(fields, 1, line),
                kind: edgeKind(fields[2] ?? '', line),
            });
            break;
        case 'instruction': {
            const instruction = new OracleInstruction(numberField(fields, 1, line));
            (currentBlock ?? panic(`an instruction outside a block: ${line}`)).instructions.push(instruction);
            currentInstruction = instruction;
            break;
        }
        case 'context':
            (currentInstruction ?? panic(`a context outside an instruction: ${line}`)).contextStore = true;
            break;
        case 'use':
            (currentInstruction ?? panic(`a use outside an instruction: ${line}`)).uses.push(placeField(fields, line));
            break;
        case 'define':
            (currentInstruction ?? panic(`a define outside an instruction: ${line}`)).defines.push(
                placeField(fields, line),
            );
            break;
        case 'terminal-use':
            (currentBlock ?? panic(`a terminal use outside a block: ${line}`)).terminalUses.push(
                placeField(fields, line),
            );
            break;
        case 'terminal-define':
            (currentBlock ?? panic(`a terminal define outside a block: ${line}`)).terminalDefines.push(
                placeField(fields, line),
            );
            break;
        case 'passes':
            run(fn, textField(fields, 1, line), line);
            current = undefined;
            break;
        default:
            panic(`an unknown record: ${line}`);
    }
}
