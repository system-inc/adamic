/*
 * A port of cohere's static_single_assignment module (static_single_assignment.go) to Adamic 0.1: the
 * graph maintenance and single static assignment form two intermediate representations share, cohere's
 * high-level IR (the React Compiler's) and Adamic's flow graph.
 *
 * It owns the algorithms and none of either IR's meaning. Reverse postorder, predecessors, evaluation
 * order, Braun's construction, redundant phi elimination and the verifier are written here once, over
 * Graph: an adapter each IR implements on its own function, block and place types. Where the two IRs
 * genuinely differ, the difference is a member of Graph rather than a fork of the algorithm:
 *
 *   - Structural fallthroughs. cohere's terminals name the block after an if or a loop beside their real
 *     edges; Adamic's have only real edges. eachEdge reports which is which.
 *   - Exception edges. Adamic's MayThrow leaves its block from the middle, at the throwing instruction;
 *     cohere emits no such terminal. eachEdge reports them as 'Exceptional', and every pass reads them as
 *     real edges today. See construct for the one place that will change (#2yz9ra9).
 *   - Context stores. A binding a closure captures is defined once in cohere, as upstream defines it;
 *     Adamic keeps captured variables out of the graph. isContextStore and contextual.
 *   - The returns place, which cohere has and Adamic does not. returns.
 *   - Nested functions, which each IR recurses into itself, around these passes.
 *
 * Where the port differs from the Go, and why:
 *
 *   - Each type here carries the house suffix (GraphInterface, EdgeType); its comment names the Go type it mirrors.
 *   - The four ids are numbers. Go's are distinct unsigned types; 0.1 has no nominal numbers, so the
 *     names say which id a number is, and nothing stops passing one for another.
 *   - Edge and Role are unions of their names, where Go's are integer enums.
 *   - A visitor renames a place by returning it, where Go's takes a pointer and writes through it. 0.1
 *     has no pointers, so the adapter writes back what the visitor returns. A visitor that renames
 *     nothing returns the place it was given.
 *   - Go's Returns is a pointer, nil for an IR with none. Here returns is the place or undefined, and
 *     setReturns writes it back.
 *   - Go's Block answers (block, found); here it is the block or undefined.
 *   - Go's Graph is a type parameter so its calls compile to direct ones. Here a pass takes the Graph
 *     interface itself, and stage 0 makes one copy of each pass per function, block and place type.
 */

/*
 * BlockIdType (Go's BlockId) identifies a basic block within one function.
 *
 * Ids are dense and assigned in creation order, which is not execution order. A function's blocks are
 * held in reverse postorder once reversePostorder has run, so a walk that wants execution order
 * iterates the array, and one that wants to address a specific block holds the id. The two are
 * deliberately different things: an index into the block array moves when the order changes, an id
 * does not.
 */
export type BlockIdType = number;

/*
 * IdentifierIdType (Go's IdentifierId) names one value.
 *
 * Two places with the same IdentifierId are the same value: not equal, the same. That is the property
 * the whole IR exists to provide, and it is what lets a later pass ask whether the object mutated here
 * is the object passed there without re-deriving aliasing from syntax.
 */
export type IdentifierIdType = number;

/*
 * DeclarationIdType (Go's DeclarationId) names one source-level binding across all the values it takes.
 *
 * One `let` reassigned three times is three IdentifierIds and one DeclarationId. Single-assignment
 * construction mints further IdentifierIds against the same DeclarationId, so a pass that wants "this
 * variable" rather than "this value" asks for the DeclarationId and keeps working afterwards.
 */
export type DeclarationIdType = number;

/*
 * EvaluationOrderType (Go's EvaluationOrder) is a monotonically increasing position in the function's evaluation.
 *
 * Terminals carry one as well as instructions, so "did A evaluate before B" is answerable across an
 * instruction and a terminal. Assigned by markEvaluationOrder in reverse postorder from 1. Zero means
 * unassigned.
 */
export type EvaluationOrderType = number;

/*
 * EdgeType (Go's Edge) is what kind of edge a successor is.
 *
 *   - 'Real' is an edge control takes.
 *   - 'Fallthrough' is a structural link, not an edge: the block after an if or a loop, which a
 *     structured terminal names whether or not control can reach it from here. It is never a
 *     predecessor. reversePostorder visits it first so loop bodies and conditional arms precede their
 *     continuation, and keeps it as an empty placeholder when nothing real reaches it.
 *   - 'Exceptional' is an edge taken when the block's last instruction throws, before it completes.
 *     Every pass treats it as 'Real' today. It is named apart because single assignment should give the
 *     handler the definitions from before the throwing instruction, not the block's end (#2yz9ra9).
 */
export type EdgeType = 'Real' | 'Fallthrough' | 'Exceptional';

// RoleType (Go's Role) says whether a place reads its value ('Use') or defines it ('Define').
export type RoleType = 'Use' | 'Define';

// PlaceVisitorType (Go's func(place *P, role Role)) sees one place and returns it, renamed or as it was.
export type PlaceVisitorType<P> = (place: P, role: RoleType) => P;

/*
 * PhiInterface (Go's Phi) is a merge: the value of place at the top of a block, given which predecessor control came from.
 *
 * operands holds one entry per entry in the block's predecessors, kept sorted by predecessor block id,
 * so iterating it is deterministic and is the order phiOperandsInOrder gives. The lookup "what came
 * from this predecessor" is phiOperand. Both fields are written in place: elimination renames a phi's
 * place and each operand's.
 */
export interface PhiInterface<P> {
    place: P;
    operands: PhiOperandInterface<P>[];
}

// PhiOperandInterface (Go's PhiOperand) is the value a phi takes when control arrives from predecessor.
export interface PhiOperandInterface<P> {
    readonly predecessor: BlockIdType;
    place: P;
}

/*
 * GraphInterface (Go's Graph) is what the passes need of an intermediate representation: its function type F, its block type
 * B and its place type P.
 *
 * An IR implements it once, and calls a pass with it and its function. The passes hold no IR type;
 * every read and write of one goes through here.
 */
export interface GraphInterface<F, B, P> {
    // entry is the block control begins at.
    readonly entry: (fn: F) => BlockIdType;
    // blockBound is one past the largest block id the function has handed out or holds.
    readonly blockBound: (fn: F) => number;
    // block is the block with an id, and undefined for one the function does not hold.
    readonly block: (fn: F, id: BlockIdType) => B | undefined;
    // blocks is the function's block array, in its current order.
    readonly blocks: (fn: F) => readonly B[];
    // setBlocks replaces the block array.
    readonly setBlocks: (fn: F, blocks: readonly B[]) => void;
    // retain forgets every block whose id keep declines, so block no longer finds it.
    readonly retain: (fn: F, keep: (id: BlockIdType) => boolean) => void;
    // placeholder replaces a block nothing real reaches but a structured terminal still names with an
    // empty unreachable block under the same id, keeping its predecessors, and returns it. Only an IR
    // that reports 'Fallthrough' edges is asked.
    readonly placeholder: (fn: F, block: B) => B;

    // id is a block's id.
    readonly id: (block: B) => BlockIdType;
    // predecessors are the blocks with a real edge into a block.
    readonly predecessors: (block: B) => readonly BlockIdType[];
    // setPredecessors replaces them.
    readonly setPredecessors: (block: B, predecessors: readonly BlockIdType[]) => void;
    // phis are a block's merges.
    readonly phis: (block: B) => readonly PhiInterface<P>[];
    // setPhis replaces them.
    readonly setPhis: (block: B, phis: readonly PhiInterface<P>[]) => void;
    // eachEdge visits a block's successors: its structural fallthrough first, if it has one, then its
    // edges in the order its terminal names them.
    readonly eachEdge: (block: B, visit: (successor: BlockIdType, edge: EdgeType) => void) => void;
    // endsInReturn reports whether a block's terminal returns from the function.
    readonly endsInReturn: (block: B) => boolean;

    // instructionCount is how many instructions a block holds.
    readonly instructionCount: (fn: F, block: B) => number;
    // eachInstructionPlace visits the places of a block's instruction at index, uses before
    // definitions, and writes back the place each visit returns, so a pass may rename them.
    readonly eachInstructionPlace: (fn: F, block: B, index: number, visit: PlaceVisitorType<P>) => void;
    // isContextStore reports whether a block's instruction at index writes a binding a closure
    // captures, which single assignment defines once rather than versioning.
    readonly isContextStore: (fn: F, block: B, index: number) => boolean;
    // contextStoreDefines reports whether a place a context store defines is the instruction's own
    // result, the one definition upstream's invariant counts. Asked only of a context store.
    readonly contextStoreDefines: (fn: F, block: B, index: number, place: P) => boolean;
    // setInstructionOrder sets the evaluation order of a block's instruction at index.
    readonly setInstructionOrder: (fn: F, block: B, index: number, order: EvaluationOrderType) => void;
    // eachTerminalPlace visits the places of a block's terminal, writing back what each visit returns.
    readonly eachTerminalPlace: (block: B, visit: PlaceVisitorType<P>) => void;
    // setTerminalOrder sets the evaluation order of a block's terminal.
    readonly setTerminalOrder: (block: B, order: EvaluationOrderType) => void;

    // params are the function's parameters, defined on entry, in an array whose elements a pass renames.
    readonly params: (fn: F) => P[];
    // returns is the place every return stores into, or undefined for an IR with none.
    readonly returns: (fn: F) => P | undefined;
    // setReturns replaces it. Asked only of an IR whose returns gave a place.
    readonly setReturns: (fn: F, place: P) => void;
    // declaration is the binding an identifier is a value of.
    readonly declaration: (fn: F, identifier: IdentifierIdType) => DeclarationIdType;
    // contextual reports whether a binding is one a nested function captures.
    readonly contextual: (fn: F, declaration: DeclarationIdType) => boolean;
    // mint makes a new value of the original's binding and returns its id.
    readonly mint: (fn: F, original: IdentifierIdType) => IdentifierIdType;
    // named reports whether a value carries a source name rather than being a temporary.
    readonly named: (fn: F, identifier: IdentifierIdType) => boolean;
    // placeString renders a value for a message.
    readonly placeString: (fn: F, identifier: IdentifierIdType) => string;

    // identifierOf is the value a place names.
    readonly identifierOf: (place: P) => IdentifierIdType;
    // withIdentifier is a place naming another value, with everything else it carries unchanged.
    readonly withIdentifier: (place: P, identifier: IdentifierIdType) => P;
}
