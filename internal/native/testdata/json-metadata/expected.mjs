// Four stage 3 root families: actual allocations, including fields outside a view.
const options = {strict:true,target:1,plugins:[{name:"x"}],hidden:"kept"};
const refs = [{path:"../a",prepend:true,circular:false}];
const info = {version:"5.8.0",root:[1,2],fileNames:["a.ts"],options,hidden:"kept"};
for (const value of [info,options,refs,[1,2],{value:7},new Map([["x",1]])]) console.log(JSON.stringify(value));
const numbers = new Array(3).fill(7);
const booleans = Object.values({a:true,b:false});
const strings = "a,b,c".split(",");
for (const value of [numbers,booleans,strings,numbers.slice(1),[...booleans,...booleans],strings.slice().reverse()]) console.log(JSON.stringify(value));
const keys = [];
const returned = [7,false,"str",null,undefined,()=>0,new Map(),[1,2],{value:7}];
class Hook {constructor(mode){this.mode=mode;} toJSON(key){keys.push(key);return returned[this.mode];}}
for(let mode=0;mode<returned.length;mode++) {
 console.log(JSON.stringify(new Hook(mode)));
 console.log(JSON.stringify({p:new Hook(mode)}));
 console.log(JSON.stringify([new Hook(mode)]));
}
console.log(JSON.stringify(keys));
// A returned object's hook does not run again at the same position.
class Returned {constructor(){this.value=7;} toJSON(){throw Error("second hook");}}
class Outer {toJSON(){return new Returned();}}
console.log(JSON.stringify(new Outer()));
console.log(JSON.stringify([0,false,undefined,NaN,null]));
console.log(JSON.stringify([1,false,2].slice(1).reverse()));
console.log(JSON.stringify([1,false,2].sort(()=>-1)));
for(const value of [[],new Array(2).fill(true),new Array(2).fill("str"),Object.entries({a:true,b:false})]) console.log(JSON.stringify(value));
const bm=new Map([[true,"str"],[false,"str"]]);
for(const value of [[...bm.keys()],[...bm.values()],[...bm],[...new Set([true,false])]]) console.log(JSON.stringify(value));
for(const value of [[1,2].map(x=>x+1),[1,2].map(x=>x>1),[1,2].map(()=>"str")]) console.log(JSON.stringify(value));
const changed=[1,2];changed[0]=9;console.log(JSON.stringify(changed));console.log(JSON.stringify(changed.splice(0,1,8)));console.log(JSON.stringify(changed));changed.fill(3);console.log(JSON.stringify(changed));
console.log(JSON.stringify([true,false,undefined]));console.log(JSON.stringify([0,undefined]));
let reentrant;class Reentrant{mode=7;toJSON(){reentrant.splice(0,1);return this.mode;}}reentrant=[new Reentrant()];console.log(JSON.stringify(reentrant));
class Derived extends Hook {constructor(){super(0);this.extra="kept";}}
console.log(JSON.stringify(new Derived()));
console.log(JSON.stringify({"10":"ten","2":"two",tail:true,"#public":"yes"}));
class Private {#secret=9;value=7;}
console.log(JSON.stringify(new Private()));
