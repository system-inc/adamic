import { testPlace, testBlock } from './core.ts';
import { panic } from 'adamic';
import type { ConstructedHIR, HIRArena, FunctionIndex, HIRFunction, PlaceInterface, Instruction, TerminalType, PatternIndex } from './core.ts';
export function placeText(place: PlaceInterface): string {
    return `${place.identifier.slot}:${place.effect}:${place.reactive ? 1 : 0}:${place.start}:${place.end}`;
}
export function dump(graph: ConstructedHIR): string { return dumpFunction(graph.arena, graph.root); }
function dumpFunction(arena: HIRArena, index: FunctionIndex): string {
    const fn = arena.read(index);
    let out = `hir-v1\nfunction ${fn.name} ${fn.kind} entry=${fn.entry.slot + 1} async=${fn.isAsync ? 1 : 0} generator=${fn.isGenerator ? 1 : 0}\n`;
    out += `params ${fn.params.map((place) => placeText(place)).join(',')}\ncontext ${fn.context.map((place) => placeText(place)).join(',')}\nreturns ${placeText(fn.returns)}\n`;
    for(const identifier of fn.identifiers) { out += `identifier ${identifier.id.slot} ${identifier.declaration.slot + 1} ${identifier.name}\n`; }
    const seen = new Map<number, boolean>();
    for(const blockId of fn.blockOrder) {
        const block = fn.block(blockId);
        out += `block ${block.id.slot + 1} ${block.kind} predecessors=${block.predecessors.map((index) => index.slot + 1).join(',')}\n`;
        for(const phi of block.phis) {
            out += `phi ${placeText(phi.place)}`;
            for(const operand of phi.operands) { out += ` ${operand.predecessor}=${placeText(operand.place)}`; }
            out += '\n';
        }
        for(const id of block.instructions) {
            const instruction = fn.instruction(id);
            seen.set(id.slot, true);
            out += instructionText(instruction, arena, fn);
        }
        out += `terminal ${block.terminalOrder} ${block.terminal.kind} ${terminalText(block.terminal,fn)}\n`;
    }
    for(const instruction of fn.instructions) { if(!seen.has(instruction.id.slot)) { out += 'orphan ' + instructionText(instruction, arena, fn); } }
    if(fn.contextDeclarations.size > 0) { out += `context-declarations [${[...fn.contextDeclarations].map((index) => index.slot + 1).sort((a, b) => a - b).join(' ')}]\n`; }
    const outlined = fn.outlined;
    if(outlined !== undefined) { const keys = [...outlined.keys()].sort((a,b) => a.slot - b.slot); for(const key of keys) { const ordinal = fn.functions.indexOf(outlined.get(key) ?? panic('missing outlined function')); if(ordinal < 0) { panic('foreign outlined function'); } out += `outlined ${key.slot} ${ordinal}\n`; } }
    for(let index = 0; index < fn.functions.length; index++) { out += `nested ${index}\n${dumpFunction(arena, fn.functions[index] ?? panic('missing nested'))}`; }
    out += 'scopes -\nend\n';
    return out;
}

// encoding/json's string alphabet, including its HTML and U+2028/U+2029 escapes.
export function quote(text: string): string {
    let out = '"';
    for(let at = 0; at < text.length; at++) {
        const code = text.charCodeAt(at);
        if(code === 34) { out += '\\"'; }
        else if(code === 92) { out += '\\\\'; }
        else if(code === 8) { out += '\\b'; }
        else if(code === 9) { out += '\\t'; }
        else if(code === 10) { out += '\\n'; }
        else if(code === 12) { out += '\\f'; }
        else if(code === 13) { out += '\\r'; }
        else if(code < 32 || code === 60 || code === 62 || code === 38 || code === 8232 || code === 8233) { out += `\\u${code.toString(16).padStart(4, '0')}`; }
        else { out += text.charAt(at); }
    }
    return out + '"';
}

function instructionText(instruction: Instruction, arena: HIRArena, fn: HIRFunction): string {
    let payload = '';
    const value = instruction.value;
    if(value.kind === 'UnsupportedNode') { payload = `{"Node":{"end":${value.nodeEnd},"kind":${quote(value.nodeKind)},"pos":${value.nodePos}},"Reason":${quote(value.reason)}}`; }
    else if(value.kind === 'Debugger') { payload = '{}'; }
    else if(value.kind === 'MetaProperty') { payload = `{"Meta":${quote(value.meta)},"Property":${quote(value.property)}}`; }
    else if(value.kind === 'TypeCastExpression') { payload = `{"Node":{"end":${value.nodeEnd},"kind":${quote('Kind' + value.nodeKind)},"pos":${value.nodePos}},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'Await') { payload = `{"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'PropertyDelete' || value.kind === 'ComputedDelete') { payload = `{"Object":${quote(placeText(value.object))},"Property":${quote(value.kind === 'PropertyDelete' ? value.property : placeText(value.property))}}`; }
    else if(value.kind === 'RegExpLiteral') { payload = `{"Flags":${quote(value.flags)},"Pattern":${quote(value.pattern)}}`; }
    else if(value.kind === 'TemplateLiteral' || value.kind === 'TaggedTemplateExpression') { const quasis = '[' + value.quasis.map((text) => quote(text)).join(',') + ']'; const subexprs = '[' + value.subexprs.map((place) => quote(placeText(place))).join(',') + ']'; payload = `{"Quasis":${quasis},"Subexprs":${subexprs}${value.kind === 'TaggedTemplateExpression' ? ',"Tag":' + quote(placeText(value.tag)) : ''}}`; }
    else if(value.kind === 'Primitive') { payload = value.literal; }
    else if(value.kind === 'LoadLocal') { payload = placeText(value.place); }
    else if(value.kind === 'UnaryExpression') { payload = `${value.operator} ${placeText(value.value)}`; }
    else if(value.kind === 'BinaryExpression') { payload = `${placeText(value.left)} ${value.operator} ${placeText(value.right)}`; }
    else if(value.kind === 'LoadGlobal') { payload = `{"BindingKind":${value.bindingKind},"Imported":${quote(value.imported)},"Name":${quote(value.name)},"Source":${quote(value.source)}}`; }
    else if(value.kind === 'StoreGlobal') { payload = `{"Name":${quote(value.name)},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'DeclareLocal' || value.kind === 'DeclareContext') { payload = `{"Kind":${value.declarationKind},"LValue":${quote(placeText(value.lvalue))}}`; }
    else if(value.kind === 'LoadContext') { payload = `{"Place":${quote(placeText(value.place))}}`; }
    else if(value.kind === 'GetIterator' || value.kind === 'NextPropertyOf') { payload = `{"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'IteratorNext') { payload = `{"Collection":${quote(placeText(value.collection))},"Iterator":${quote(placeText(value.iterator))}}`; }
    else if(value.kind === 'ArrayExpression') { payload = `{"Elements":[${value.elements.map((element) => `{"Hole":${element.hole ? 'true' : 'false'},"Place":${quote(placeText(element.place))},"Spread":${element.spread ? 'true' : 'false'}}`).join(',')}]}`; }
    else if(value.kind === 'ObjectExpression') { payload = `{"Properties":[${value.properties.map((property) => `{"ComputedKey":${property.computedKey === undefined ? 'null' : quote(placeText(property.computedKey))},"Key":${quote(property.key)},"Spread":${property.spread ? 'true' : 'false'},"Value":${quote(placeText(property.value))}}`).join(',')}]}`; }
    else if(value.kind === 'JsxText') { payload = `{"Value":${quote(value.text)}}`; }
    else if(value.kind === 'JsxExpression' || value.kind === 'JsxFragment') {
        const children = '[' + value.children.map((place) => quote(placeText(place))).join(',') + ']';
        if(value.kind === 'JsxFragment') { payload = `{"Children":${children}}`; }
        else {
            const props = '[' + value.props.map((prop) => `{"Name":${quote(prop.name)},"Spread":${prop.spread ? 'true' : 'false'},"Value":${quote(placeText(prop.value))}}`).join(',') + ']';
            const tag = `{"Name":${quote(value.tag.name)},"Place":${value.tag.place === undefined ? 'null' : quote(placeText(value.tag.place))}}`;
            payload = `{"Children":${children},"Props":${props},"Tag":${tag}}`;
        }
    }
    else if(value.kind === 'Destructure') { const pattern = patternText(fn,value.pattern); payload = `{"Kind":${value.declarationKind},"LValue":${patternText(fn,value.lvaluePattern)},"Pattern":${pattern},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'ObjectMethod') { arena.read(value.functionReference.index); if(fn.functions[value.functionReference.ordinal] !== value.functionReference.index) { panic('method function index disagrees'); } payload = `{"Function":${value.functionReference.ordinal},"Key":${quote(value.key)}}`; }
    else if(value.kind === 'FunctionExpression') { arena.read(value.functionReference.index); if(fn.functions[value.functionReference.ordinal] !== value.functionReference.index) { panic('function reference ordinal disagrees with arena index'); } payload = `{"Captures":[${value.captures.map((place) => quote(placeText(place))).join(',')}],"Function":${value.functionReference.ordinal}}`; }
    else if(value.kind === 'StoreLocal' || value.kind === 'StoreContext') { payload = `{"Kind":${value.declarationKind},"LValue":${quote(placeText(value.lvalue))},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'PrefixUpdate' || value.kind === 'PostfixUpdate') { payload = `{"LValue":${quote(placeText(value.lvalue))},"Operation":${quote(value.operation)},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'PropertyLoad') { payload = `{"Object":${quote(placeText(value.object))},"Optional":${value.optional ? 'true' : 'false'},"Property":${quote(value.property)}}`; }
    else if(value.kind === 'ComputedLoad') { payload = `{"Object":${quote(placeText(value.object))},"Optional":${value.optional ? 'true' : 'false'},"Property":${quote(placeText(value.property))}}`; }
    else if(value.kind === 'PropertyStore') { payload = `{"Object":${quote(placeText(value.object))},"Property":${quote(value.property)},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'ComputedStore') { payload = `{"Object":${quote(placeText(value.object))},"Property":${quote(placeText(value.property))},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'CallExpression' || value.kind === 'MethodCall' || value.kind === 'NewExpression') {
        const args = '[' + value.args.map((argument) => `{"Place":${quote(placeText(argument.place))},"Spread":${argument.spread ? 'true' : 'false'}}`).join(',') + ']';
        if(value.kind === 'NewExpression') { payload = `{"Args":${args},"Callee":${quote(placeText(value.callee))}}`; }
        else {
            const origin = `{"Export":${quote(value.origin.exported)},"Module":${quote(value.origin.module)}}`;
            if(value.kind === 'CallExpression') { payload = `{"Args":${args},"Callee":${quote(placeText(value.callee))},"CalleeOrigin":${origin},"Optional":${value.optional ? 'true' : 'false'}}`; }
            else { payload = `{"Args":${args},"CalleeOrigin":${origin},"Optional":${value.optional ? 'true' : 'false'},"Property":${quote(placeText(value.property))},"Receiver":${quote(placeText(value.receiver))}}`; }
        }
    }
    else if(value.kind === 'StartMemoize') { const deps = value.deps === undefined ? 'null' : '[' + value.deps.map((dep) => `{"Path":[${dep.path.map((entry) => `{"Optional":${entry.optional ? 'true' : 'false'},"Property":${quote(entry.property)}}`).join(',')}],"Root":{"IsGlobal":${dep.root.isGlobal ? 'true' : 'false'},"Name":${quote(dep.root.name)},"Place":${quote(placeText(dep.root.place))}}}`).join(',') + ']'; payload = `{"Deps":${deps},"ManualMemoId":${value.manualMemoId}}`; }
    else if(value.kind === 'FinishMemoize') { payload = `{"ManualMemoId":${value.manualMemoId},"Pruned":${value.pruned ? 'true' : 'false'},"Value":${quote(placeText(value.value))}}`; }
    return `instruction ${instruction.id.slot} ${instruction.order} ${placeText(instruction.lvalue)} ${instruction.start}:${instruction.end} ${instruction.value.kind} ${payload}\n`;
}

export function terminalText(terminal: TerminalType,fn: HIRFunction | undefined = undefined): string {
    if(terminal.kind === 'Return') { return placeText(terminal.value ?? panic('missing terminal value')); }
    if(terminal.kind === 'Throw') { return `{"Value":${quote(placeText(terminal.value ?? panic('missing terminal value')))}}`; }
    if(terminal.kind === 'Goto') { return `{"Block":${(terminal.block ?? panic('missing terminal index')).slot + 1},"Variant":${terminal.variant}}`; }
    if(terminal.kind === 'If' || terminal.kind === 'Branch') { return `{"Alternate":${(terminal.alternate ?? panic('missing terminal index')).slot + 1},"Consequent":${(terminal.consequent ?? panic('missing terminal index')).slot + 1},"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1},"Test":${quote(placeText(testPlace(terminal)))}}`; }
    // gap 1 (GAPS.md): narrow the flag record, then read its required boolean.
    if(terminal.kind === 'Optional') { const flag = terminal.optionalFlag ?? panic('missing optional presence/value'); if(!flag.present) { panic('Optional terminal requires its boolean'); } return `{"Fallthrough":${(terminal.fallthrough ?? panic('missing optional exit')).slot + 1},"Optional":${flag.value ? 'true' : 'false'},"Test":${testBlock(terminal).slot + 1}}`; }
    if(terminal.kind === 'Logical') { return `{"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1},"Operator":${quote(terminal.operator ?? panic('missing logical operator'))},"Test":${(terminal.testBlock ?? panic('missing terminal index')).slot + 1}}`; }
    if(terminal.kind === 'Ternary') { return `{"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1},"Test":${(terminal.testBlock ?? panic('missing terminal index')).slot + 1}}`; }
    if(terminal.kind === 'While' || terminal.kind === 'DoWhile') { return `{"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1},"Loop":${(terminal.loop ?? panic('missing terminal index')).slot + 1},"Test":${(terminal.testBlock ?? panic('missing terminal index')).slot + 1}}`; }
    if(terminal.kind === 'For') { return `{"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1},"Init":${(terminal.init ?? panic('missing terminal index')).slot + 1},"Loop":${(terminal.loop ?? panic('missing terminal index')).slot + 1},"Test":${(terminal.testBlock ?? panic('missing terminal index')).slot + 1},"Update":${terminal.update === undefined ? 0 : terminal.update.slot + 1}}`; }
    if(terminal.kind === 'ForOf') { return `{"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1},"Init":${(terminal.init ?? panic('missing terminal index')).slot + 1},"Loop":${(terminal.loop ?? panic('missing terminal index')).slot + 1},"Test":${(terminal.testBlock ?? panic('missing terminal index')).slot + 1}}`; }
    if(terminal.kind === 'ForIn') { return `{"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1},"Init":${(terminal.init ?? panic('missing terminal index')).slot + 1},"Loop":${(terminal.loop ?? panic('missing terminal index')).slot + 1}}`; }
    if(terminal.kind === 'Scope') { return `{"Block":${(terminal.block ?? panic('missing scope body')).slot + 1},"Fallthrough":${(terminal.fallthrough ?? panic('missing scope exit')).slot + 1},"Scope":${(fn ?? panic('Scope dump requires its arena')).scopeId(terminal.scope ?? panic('missing scope index'))}}`; }
    if(terminal.kind === 'Sequence') { return `{"Block":${(terminal.block ?? panic('missing sequence body')).slot + 1},"Fallthrough":${(terminal.fallthrough ?? panic('missing sequence exit')).slot + 1}}`; }
    if(terminal.kind === 'MaybeThrow') { return `{"Continuation":${(terminal.continuation ?? panic('missing continuation')).slot + 1},"Handler":${terminal.handler === undefined ? 0 : terminal.handler.slot + 1}}`; }
    if(terminal.kind === 'Label') { return `{"Block":${(terminal.block ?? panic('missing terminal index')).slot + 1},"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1}}`; }
    if(terminal.kind === 'Try') { return `{"Block":${(terminal.block ?? panic('missing terminal index')).slot + 1},"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1},"Handler":${(terminal.handler ?? panic('missing terminal index')).slot + 1},"HandlerBinding":${terminal.handlerBinding === undefined ? 'null' : quote(placeText(terminal.handlerBinding))}}`; }
    if(terminal.kind === 'Switch') { const cases = (terminal.cases ?? panic('missing cases')).map((clause) => `{"Block":${clause.block.slot + 1},"Test":${clause.test === undefined ? 'null' : quote(placeText(clause.test))}}`).join(','); return `{"Cases":[${cases}],"Fallthrough":${(terminal.fallthrough ?? panic('missing terminal index')).slot + 1},"Test":${quote(placeText(testPlace(terminal)))}}`; }
    return '{}';
}

function patternText(fn: HIRFunction, index: PatternIndex): string {
 const value = fn.pattern(index);
 if(value.kind === 'Place') { return `{"Place":${quote(placeText(value.place))}}`; }
 const rest = value.rest === undefined ? 'null' : quote(placeText(value.rest));
 if(value.kind === 'Object') { return `{"Properties":[${value.properties.map((item) => `{"ComputedKey":${item.computedKey === undefined ? 'null' : quote(placeText(item.computedKey))},"Default":${item.defaultValue === undefined ? 'null' : quote(placeText(item.defaultValue))},"Key":${quote(item.key)},"Value":${patternText(fn,item.value)}}`).join(',')}],"Rest":${rest}}`; }
 return `{"Elements":[${value.elements.map((item) => { const nested = item.value; return `{"Default":${item.defaultValue === undefined ? 'null' : quote(placeText(item.defaultValue))},"Value":${nested === undefined ? 'null' : patternText(fn,nested)}}`; }).join(',')}],"Rest":${rest}}`;
}
