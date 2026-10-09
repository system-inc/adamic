'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),assert=require('node:assert/strict');
const {nodes,ts,census}=require('./census.cjs');
const [before,after,output,selected]=process.argv.slice(2);
assert(before&&after&&output,'usage: verify.cjs BEFORE AFTER REPORT [case]');
const coreNames=['mapIterator','flatMapIterator','mapDefinedIterator','singleIterator','arrayReverseIterator','createSet'];
const checkerNames=['generateJsxAttributes','generateJsxChildren','generateLimitedTupleElements','generateObjectLiteralElements','getUnmatchedProperties'];
function load(tree) {
    const declarations=[];
    for(const file of ['core.ts','checker.ts']) {
        const source=ts.createSourceFile(file,fs.readFileSync(path.join(tree,'src/compiler',file),'utf8'),99,true);
        for(const name of [...(file==='core.ts'?coreNames:checkerNames),'temporaryExplicitIterator','temporaryIteratorClose','temporaryIteratorResult','temporaryFlatMapIterator']) {
            const found=nodes(source,n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);
            if(!found.length)continue;
            assert.equal(found.length,1,name);
            declarations.push(found[0].getText(source).replace(/^export /,''));
        }
        if(file==='checker.ts') {
            const anonymous=nodes(source,n=>ts.isFunctionExpression(n)&&!n.name&&n.body.getText(source).includes('value: elem')||ts.isFunctionExpression(n)&&!!n.asteriskToken);
            assert.equal(anonymous.length,1,'anonymous elaboration');
            declarations.push('function anonymous(elem: unknown) { return ('+anonymous[0].getText(source)+')(); }');
        }
    }
    return ts.transpileModule(declarations.join('\n'),{compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.None}}).outputText;
}
const versions=[load(before),load(after)];
const results=[];
function compare(name,probe) {
    if(selected&&selected!==name)return;
    const observed=versions.map(code=>{
        const trace=[];
        const note=(name,value)=>{trace.push(name);return value;};
        const context=vm.createContext({Symbol,TypeError,
            isArray:Array.isArray,contains:(array,value,equals)=>array.some(x=>equals(x,value)),arrayFrom:Array.from,unorderedRemoveItemAt:(array,i)=>{array[i]=array[array.length-1];array.pop();},
            length:array=>note('length',array?.length||0),
            isJsxSpreadAttribute:prop=>note('spread?',prop.kind==='spread'),
            isHyphenatedJsxName:name=>note('hyphen?',name.includes('-')),
            getTextOfJsxAttributeName:name=>note('text',name),
            getStringLiteralType:name=>note('stringType',{text:name}),
            getNumberLiteralType:number=>note('numberType',{number}),
            getElaborationElementForJsxChild:(child,type)=>note('child',child.skip?undefined:{errorNode:child,innerExpression:child,nameType:type}),
            isTupleLikeType:target=>note('tuple?',target.tuple),getPropertyOfType:(type,name)=>note('property:'+name,type.props?.[name]),
            isOmittedExpression:node=>note('omitted?',node.skip),getEffectiveCheckNode:node=>note('checkNode',node),
            isSpreadAssignment:node=>note('spread?',node.kind==='spread'),
            getSymbolOfDeclaration:node=>note('symbol',node),getLiteralTypeFromProperty:node=>note('literalType',node.type),
            TypeFlags:{Never:1,StringOrNumberLiteralOrUnique:2,Unit:4,Any:8},
            SyntaxKind:{SetAccessor:10,GetAccessor:11,MethodDeclaration:12,ShorthandPropertyAssignment:13,PropertyAssignment:14},
            Debug:{assertNever:()=>{throw new Error('assertNever');}},isComputedNonLiteralName:name=>note('computed?',name==='computed'),Diagnostics:{Type_of_computed_property_s_value_is_0_which_is_not_assignable_to_type_1:'computed'},
            getPropertiesOfType:type=>note('properties',type.properties),isStaticPrivateIdentifierProperty:prop=>note('private?',prop.private),SymbolFlags:{Optional:1},CheckFlags:{Partial:2},getCheckFlags:prop=>note('checkFlags',prop.checkFlags||0),getTypeOfSymbol:prop=>note('symbolType',prop.type),getRegularTypeOfLiteralType:type=>note('regularType',type),
        });
        vm.runInContext(code,context);
        const value=probe(context,trace,note);
        return JSON.parse(JSON.stringify({trace,value}));
    });
    assert.deepEqual(observed[1],observed[0],name+' changed values or evaluation timing');
    results.push({name,status:'pass',...observed[1]});
}
function source(trace,values,custom={}) {
    let index=0;
    return {[Symbol.iterator](){trace.push('iterator');return {
        get next(){trace.push('next-get');return function(){trace.push('next');return index<values.length?{done:false,value:values[index++]}:{done:true,value:undefined};};},
        get return(){trace.push('return-get');return function(...args){trace.push('return:'+args.length);return {done:true,value:undefined};};},...custom,
    };}};
}
function drain(iter) {const result=[];for(let i=0;i<20;i++){const value=iter.next();result.push(value);if(value.done)return result;}throw new Error('iterator did not finish');}
for(const name of ['mapIterator','mapDefinedIterator','flatMapIterator'])compare(name,(c,trace)=>{
    const input=source(trace,[1,2,3]);
    const callback=x=>{trace.push('map:'+x);return name==='flatMapIterator'?(x===2?undefined:[x,x+10]):name==='mapDefinedIterator'&&x===2?undefined:x*2;};
    const iter=c[name](input,callback);assert.deepEqual(trace,[],'laziness at construction');
    const values=[iter.next('ignored')];
    values.push(iter.next('sent'));
    values.push(iter.return(undefined),iter.next());
    const fresh=c[name](source(trace,[1,2,3]),callback);
    values.push(...drain(fresh));
    const unused=c[name](source(trace,[1]),callback);values.push(unused.return(undefined),unused.next());
    const throwing=c[name](source(trace,[1]),()=>{throw new Error('mapped');});
    try{throwing.next();}catch(e){values.push(e.message);}
    values.push(throwing.next());
    const cancelled=c[name](source(trace,[1]),callback);cancelled.next();
    const failure={tag:'throw'};try{cancelled.throw(failure);}catch(e){if(name==='flatMapIterator') values.push({name:e.name,message:e.message}); else {assert.equal(e,failure);values.push(e);}}
    values.push(cancelled.next());
    return values;
});
compare('delegation',(c,trace)=>{
    const outer=source(trace,[1,2]);
    const iter=c.flatMapIterator(outer,x=>({[Symbol.iterator](){trace.push('inner:'+x);let calls=0;return {
        next(value){trace.push('inner-next:'+String(value));return calls++===0?{done:false,value:x}:{done:true,value:undefined};},
        return(value){trace.push('inner-return:'+String(value));return {done:false,value:90};},
        throw(value){trace.push('inner-throw:'+value);return {done:true,value:undefined};},
    };}}));
    assert.deepEqual(trace,[],'delegation is eager');
    return [iter.next(999),iter.return(undefined),iter.next(7),iter.throw('recover'),...drain(iter)];
});
compare('delegated-result',(c,trace)=>{
    const input=source(trace,[1]);
    let step=0;
    const inner={ [Symbol.iterator](){return {
        next(){const done=step++>0;return {
            get done(){trace.push('delegated-done');return done;},
            get value(){trace.push('delegated-value');return done?undefined:41;},
        };},
    };}};
    const iter=c.flatMapIterator(input,()=>inner);
    const first=iter.next();
    // A yield* returns the delegate's result object without reading its value.
    const atYield=[...trace];
    const value=first.value;
    const final=iter.next();
    return {atYield,value,final};
});
compare('value-getter-throw',(c,trace)=>{
    const failure={tag:'getter'};
    const input={ [Symbol.iterator](){return {
        next(){return {done:false,get value(){trace.push('value-get');throw failure;}};},
        return(){trace.push('closed');return {done:true,value:undefined};},
    };}};
    const result=[];
    for(const name of ['mapIterator','mapDefinedIterator','flatMapIterator']) {
        const iter=c[name](input,x=>x);
        try{iter.next();}catch(e){assert.equal(e,failure);result.push(e);}
        result.push(iter.next());
    }
    return result;
});
compare('singleIterator',(c,trace)=>{
    const iter=c.singleIterator(17);assert.deepEqual(trace,[]);assert.equal(iter[Symbol.iterator](),iter);return [...drain(iter),iter.return(undefined),iter.next()];
});
compare('arrayReverseIterator',(c,trace)=>{
    const array=new Proxy([1,2,3],{get(target,key,receiver){trace.push('read:'+String(key));return Reflect.get(target,key,receiver);}});
    const iter=c.arrayReverseIterator(array);assert.deepEqual(trace,[],'laziness: array length read before next');
    array.push(4);const first=iter.next();array[2]=9;return [first,...drain(iter)];
});
compare('set',(c,trace)=>{
    const set=c.createSet(value=>value%2,(a,b)=>a===b);set.add(1).add(3).add(2);
    const iter=set.values();set.add(5);const first=iter.next();set.add(7);const values=[first,...drain(iter)];
    const entries=set.entries();entries.next();entries.return(undefined);values.push(entries.next());
    return {values,entries:Array.from(set.entries()),keys:Array.from(set.keys()),all:Array.from(set)};
});
for(const name of checkerNames)compare(name,(c,trace)=>{
    const prop=(kind,name,extra={})=>({kind,name,initializer:{text:name},...extra});
    let args;
    if(name==='generateJsxAttributes')args=[{properties:[prop('spread','skip'),prop('attr','a-b'),prop('attr','ok'),prop('attr','last')]}];
    if(name==='generateJsxChildren')args=[{children:[{skip:true},{text:'first'},{text:'second'}]},()=>({message:'invalid'})];
    if(name==='generateLimitedTupleElements')args=[{elements:[{skip:true},{text:'first'},{text:'second'}]},{tuple:true,props:{'0':{},'1':{}}}];
    if(name==='generateObjectLiteralElements')args=[{properties:[prop('spread','skip'),prop(14,'never',{type:{flags:1}}),...[10,11,12,13,14].map(kind=>prop(kind,'name'+kind,{type:{flags:2}}))]}];
    if(name==='getUnmatchedProperties')args=[{props:{matched:{type:{flags:8}}}},{properties:[{escapedName:'private',private:true},{escapedName:'optional',flags:1},{escapedName:'missing',flags:0},{escapedName:'matched',flags:0,type:{flags:4}}]},false,true];
    const iter=c[name](...args);assert.deepEqual(trace,[],'checker is eager');
    const first=iter.next();const cancelled=c[name](...args);cancelled.return(undefined);
    return [first,...drain(iter),cancelled.next()];
});
compare('anonymous',(c,trace)=>{const iter=c.anonymous({errorNode:'one'});assert.deepEqual(trace,[]);return drain(iter);});
if(!selected)assert.equal(census(path.resolve(after)).generatorCount,0,'unrewritten generator');
fs.writeFileSync(output,JSON.stringify({status:'pass',cases:results},null,2)+'\n');
console.log(JSON.stringify({status:'pass',cases:results.map(r=>r.name)}));
