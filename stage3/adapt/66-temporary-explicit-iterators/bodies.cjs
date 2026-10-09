'use strict';
module.exports = {
mapIterator: `
    let input: Iterator<T, unknown, unknown> | undefined;
    let advance: (() => IteratorResult<T, unknown>) | undefined;
    return temporaryExplicitIterator<U>(() => {
        if (!input) {
            input = iter[Symbol.iterator]();
            const iterator = input, next = iterator.next;
            advance = () => next.call(iterator);
        }
        const result = temporaryIteratorResult(advance!());
        if (result.done) return { done: true, value: undefined };
        const x = result.value;
        try { return { done: false, value: mapFn(x) }; }
        catch (error) { try { temporaryIteratorClose(input); } catch {} throw error; }
    }, () => temporaryIteratorClose(input));
`,
mapDefinedIterator: `
    let input: Iterator<T, unknown, unknown> | undefined;
    let advance: (() => IteratorResult<T, unknown>) | undefined;
    return temporaryExplicitIterator<U & ({} | null)>(() => {
        if (!input) {
            input = iter[Symbol.iterator]();
            const iterator = input, next = iterator.next;
            advance = () => next.call(iterator);
        }
        while (true) {
            const result = temporaryIteratorResult(advance!());
            if (result.done) return { done: true, value: undefined };
            const x = result.value;
            try {
                const value = mapFn(x);
                if (value !== undefined) return { done: false, value };
            }
            catch (error) { try { temporaryIteratorClose(input); } catch {} throw error; }
        }
    }, () => temporaryIteratorClose(input));
`,
singleIterator: `
    let pending = true;
    return temporaryExplicitIterator<T>(() => {
        if (!pending) return { done: true, value: undefined };
        pending = false;
        return { done: false, value };
    });
`,
arrayReverseIterator: `
    let index: number | undefined;
    return temporaryExplicitIterator<T>(() => {
        if (index === undefined) index = array.length - 1;
        if (index < 0) return { done: true, value: undefined };
        const value = array[index]!;
        index--;
        return { done: false, value };
    });
`,
flatMapIterator: `
    let input: Iterator<T, unknown, unknown> | undefined;
    let advance: (() => IteratorResult<T, unknown>) | undefined;
    let completed = false;
    let delegated: Iterator<U, undefined, unknown> | undefined;
    let delegateNext: ((value: unknown) => IteratorResult<U, undefined>) | undefined;
    function pull(value: unknown): IteratorResult<U, undefined> {
        if (!input) {
            input = iter[Symbol.iterator]();
            const iterator = input, next = iterator.next;
            advance = () => next.call(iterator);
        }
        while (true) {
            if (delegated) {
                try {
                    const result = temporaryIteratorResult(delegateNext!(value));
                    if (!result.done) return result;
                    void result.value;
                    delegated = undefined;
                }
                catch (error) { try { temporaryIteratorClose(input); } catch {} throw error; }
            }
            const result = temporaryIteratorResult(advance!());
            if (result.done) { completed = true; return { done: true, value: undefined }; }
            const x = result.value;
            try {
                const inner = mapfn(x);
                if (!inner) continue;
                delegated = inner[Symbol.iterator]();
                const iterator = delegated, next = iterator.next;
                delegateNext = value => next.call(iterator, value);
                value = undefined;
            }
            catch (error) { try { temporaryIteratorClose(input); } catch {} throw error; }
        }
    }
    return temporaryExplicitIterator<U>(pull, () => temporaryIteratorClose(input), ({ kind, value }) => {
        let completion = kind === "return" ? value : undefined;
        try {
            if (delegated) {
                if (kind === "return") {
                    const handler = delegated.return;
                    if (handler != null) {
                        const result = temporaryIteratorResult<U, undefined>(handler.call(delegated, completion));
                        if (!result.done) return result;
                        completion = result.value;
                    }
                }
                else {
                    const handler = delegated.throw;
                    if (handler == null) {
                        temporaryIteratorClose(delegated);
                        throw new TypeError("The iterator does not provide a 'throw' method.");
                    }
                    const result = temporaryIteratorResult<U, undefined>(handler.call(delegated, value));
                    if (!result.done) return result;
                    void result.value;
                    delegated = undefined;
                    return pull(undefined);
                }
            }
            if (kind === "throw") throw value;
        }
        catch (error) { try { temporaryIteratorClose(input); } catch {} throw error; }
        temporaryIteratorClose(input);
        completed = true;
        return { done: true, value: completion };
    }, () => completed);
`,
getElementIterator: `
        let input: Generator<TElement, undefined, unknown> | undefined;
        return temporaryExplicitIterator<TElement>(value => {
            input ??= temporaryFlatMapIterator(multiMap.values(), value => isArray(value) ? value : singleIterator(value));
            const result = input.next(value);
            return result.done ? { done: true, value: undefined } : result;
        }, () => { input?.return(undefined); }, ({ kind, value }) => kind === "throw" ? input!.throw(value) : input!.return(value));
`,
entries: `
            let input: Iterator<TElement, unknown, unknown> | undefined;
            return temporaryExplicitIterator<[TElement, TElement]>(() => {
                input ??= getElementIterator();
                const result = input.next();
                return result.done ? { done: true, value: undefined } : { done: false, value: [result.value, result.value] };
            }, () => temporaryIteratorClose(input));
`,
};
const fs = require('node:fs');
const {ts,nodes} = require('./census.cjs');
const reviewed = JSON.parse(fs.readFileSync(require('node:path').join(__dirname,'reviewed.json'),'utf8'));
const elaboration = 'ElaborationIterator extends IterableIterator<infer Element> ? Element : never';
function yieldReturns(text) {
    const prefix = 'function* f() ';
    const source=ts.createSourceFile('body.ts',prefix+text,99,true);
    const yields=nodes(source,ts.isYieldExpression);
    for(const node of yields.sort((a,b)=>b.getStart(source)-a.getStart(source))) {
        const statement=node.parent;
        if(!ts.isExpressionStatement(statement))throw new Error('unreviewed yield form');
        text=text.slice(0,statement.getStart(source)-prefix.length)+'return { done: false, value: '+node.expression.getText(source)+' };'+text.slice(statement.end-prefix.length);
    }
    return text;
}
for(const key of ['generateJsxAttributes','generateObjectLiteralElements','getUnmatchedProperties']) {
    const row=reviewed.find(row=>row.name===key);
    const source=ts.createSourceFile('body.ts',row.body,99,true);
    const loop=nodes(source,ts.isForOfStatement)[0];
    const parameter=loop.initializer.declarations[0].name.getText(source);
    const iterable=loop.expression.getText(source);
    const prelude=row.body.slice(1,loop.getStart(source));
    const item=key==='getUnmatchedProperties'?'Symbol':'(typeof node.properties)[number]';
    const yielded=key==='getUnmatchedProperties'?'Symbol':elaboration;
    const statements=yieldReturns(loop.statement.getText(source)).slice(1,-1);
    module.exports[key]=`
        let initialized = false;
        let input: Iterator<${item}, unknown, unknown> | undefined;
        let advance: (() => IteratorResult<${item}, unknown>) | undefined;
        return temporaryExplicitIterator<${yielded}>(() => {
            if (!initialized) {
                initialized = true;
                function initialize(): Iterator<${item}, unknown, unknown> | undefined {
                    ${prelude}
                    return ${iterable}[globalThis.Symbol.iterator]();
                }
                input = initialize();
                if (input) {
                    const iterator = input, next = iterator.next;
                    advance = () => next.call(iterator);
                }
            }
            if (!advance) return { done: true, value: undefined };
            while (true) {
                const result = temporaryIteratorResult(advance());
                if (result.done) return { done: true, value: undefined };
                const ${parameter} = result.value;
                try { ${statements} }
                catch (error) { try { temporaryIteratorClose(input); } catch {} throw error; }
            }
        }, () => temporaryIteratorClose(input));
    `;
}
module.exports.generateJsxChildren = `
        let initialized = false;
        let index = 0;
        let memberOffset = 0;
        return temporaryExplicitIterator<${elaboration}>(() => {
            if (!initialized) {
                initialized = true;
                if (!length(node.children)) return { done: true, value: undefined };
            }
            while (index < node.children.length) {
                const i = index++;
                const child = node.children[i]!;
                const nameType = getNumberLiteralType(i - memberOffset);
                const elem = getElaborationElementForJsxChild(child, nameType, getInvalidTextDiagnostic);
                if (elem) return { done: false, value: elem };
                memberOffset++;
            }
            return { done: true, value: undefined };
        });
`;
module.exports.generateLimitedTupleElements = `
        let len: number | undefined;
        let index = 0;
        return temporaryExplicitIterator<${elaboration}>(() => {
            if (len === undefined) len = length(node.elements);
            while (index < len) {
                const i = index++;
                if (isTupleLikeType(target) && !getPropertyOfType(target, ("" + i) as __String)) continue;
                const elem = node.elements[i]!;
                if (isOmittedExpression(elem)) continue;
                const nameType = getNumberLiteralType(i);
                const checkNode = getEffectiveCheckNode(elem);
                return { done: false, value: { errorNode: checkNode, innerExpression: checkNode, nameType } };
            }
            return { done: true, value: undefined };
        });
`;
module.exports['<anonymous>'] = `
                                let pending = true;
                                return temporaryExplicitIterator<typeof elem>(() => {
                                    if (!pending) return { done: true, value: undefined };
                                    pending = false;
                                    return { done: false, value: elem };
                                });
                            `;

module.exports.flatMapRuntime = `function temporaryFlatMapIterator<T, U>(iter: Iterable<T>, mapfn: (x: T) => readonly U[] | Iterable<U> | undefined): Generator<U, undefined, unknown> {${module.exports.flatMapIterator}}
`;
module.exports.flatMapIterator = `
    return temporaryFlatMapIterator(iter, mapfn);
`;
