package typeaware

import "testing"

var volumeRules = []string{"no_unsafe_type_assertion", "no_unsafe_member_access", "prefer_nullish_coalescing", "no_shadow", "no_unsafe_enum_comparison", "no_unsafe_assignment", "no_confusing_void_expression", "consistent_return", "switch_exhaustiveness_check", "unbound_method"}

func volumeOracle(h *harness, name, source string) string {
	return typeAwareOverlayOracle(h, name+"-build", name, "adamic_"+name+".go", source, name+"-overlay.json", false)
}

// These controls exercise decisions, report spans and repairs, not implementation structure.
func volumeControls() []string {
	return []string{
		"declare const x: unknown; const y=x as number;",
		"const x={a:1}; const y=x as {a:number;b:number};",
		"const o={foo:'hi',bar:1} as {foo:string}; const p={foo:'hi'} as {foo:string};",
		"const a=new Map /* comment */ ( ) as Map<string,string>; const b=new Map() as Map<string,string>;",
		"declare const x:any; const y=x as string; const z='x' as any;",
		"function f<T>(x:unknown) { return x as T; } function g<T extends {a:number}>(x:{a:number}) { return x as T; }",
		"declare const x:any; x.a.b; x['c']; declare const y:{[key:string]:number}; declare const key:any;y[key];",
		"declare const a:number|null; const b=a || 0; const c=a === null ? 0 : a;",
		"declare const a:string|undefined; if(a === undefined) { a = '🌍'; }",
		"declare const a:number|null; const b= a === null ? (true ? 0 : 1) : a;",
		"let a:string|undefined; if(a == null) { /* 世界 */ a = 'x'; // 🌍\n }",
		"enum E { A, B } enum F { A } declare const e:E; e===F.A; e===0; e===E.B; E.A===E.B;",
		"enum E {A='a',B='b'} declare const e:E; e==='a'; switch(e) { case 'a': break; }",
		"declare const x:any; const y:string=x; const a:[any,number]=[x,1];const [b,c]=a;",
		"declare const x:{a:any}; const {a}=x; declare const y:{p:{v:any}}; const {p:{v}}=y;",
		"declare const x:{a:{b:any}}; const {a:{b}}=x; declare const z:{a:number}; const {a:c}=z;",
		"declare const x:Set<any>; const a:{p:Set<string>}={p:x};",
		"declare const x:number; function f(x:number) { return x; }",
		"const x=1; (()=>{const x=2;return x;})();const g=(x:number)=>x;",
		"type T=number; class A<T> { m<T>(x:T) {return x;} static n<T>(){} } interface B<T>{p:T};",
		"enum E { E } const a=function a(){}; const b=wrap(function b(){}); declare function wrap(x:unknown):unknown;",
		"/* 世界 🌍 */\r\nconst é=1; function f(é:number) { return é; }",
		"export function f(){}; function g(){function f(){};return f;} export const x=1; function h(x:number){};",
		"declare function v():void; function f(x:boolean) { if(x) return v(); return; }",
		"declare function v():void; const f=()=>v(); function g(){return v();} function h(x:boolean){if(x)return v();console.log('x');}",
		"declare function v():void; const a= [v(), true?v():v()]; const f=()=>{v();};",
		"function f(x:boolean):number { if(x) return 1; } function g(x:boolean):number { if(x) return 1; return; }",
		"function f():boolean; function f(x:boolean):void; function f(x?:boolean) {if(x)return;return true;}",
		"async function f(x:boolean):Promise<void> { if(x) return;return Promise.resolve(); }",
		"async function f(x:boolean):Promise<Promise<void|number>> { if(x)return;return 1; }",
		"type P<T>={then(cb:(x:T)=>any):any}; async function f(x:boolean):P<void> {if(x)return;return {} as P<void>;}",
		"type P<T>={then(cb:number):any}; async function f(x:boolean):P<void> {if(x)return;return {} as P<void>;}",
		"type P<T>={then: number}; async function f(x:boolean):P<void> {if(x)return;return {} as P<void>;}",
		"function f(x:boolean) { while(true) { if(x)return 1; break; } }",
		"function f(x:boolean) {while(0xbn){if(x)return 1;}} function g(x:boolean){while(0x0n){if(x)return 1;}} function é(x:boolean){if(x)return 1;}",
		"function f(x:boolean) { outer: while(true) { switch(x){case true:break outer;default:return 1;} } }",
		"function f(x:boolean) { do { if(x)return 1;continue; } while(false); }",
		"function f(x:boolean) { for(;x;) { return 1; } } function g(){for(;;){return 1;}}",
		"function f(x:boolean) { for(;(x);) {return 1;} } function g() {try{return 1;}catch{}}",
		"function f(x:boolean) { switch(x){case true:return 1;default:return 2;} }",
		"function f(x:boolean) { try { return 1; } catch { throw 1; } finally { console.log('x'); } }",
		"declare const x:'a'|'b'|null; switch(x){case 'a':break;} declare const s:unique symbol; declare const y:typeof s|undefined;switch(y){case undefined:break;}",
		"declare const x:boolean; switch(x){case true:break;case false:break;} declare const y:number;switch(y){}",
		"class A { m(){}; f=function(){}; a=()=>{}; v(this:void){}; t(this:A){} };const a=new A;const m=a.m;const f=a.f;const b=a.a;const v=a.v;const t=a.t;a.m();if(a.m){};!a.m;",
		"interface A {m():void}; declare const a:A;const {m}=a;const {m:n}=a;const f=a['m'];",
		"interface A {m():void};declare const a:A;let m:()=>void;({m}=a);const f={m:function(){}}.m;",
		"declare const a:{m():void}|{m(this:void):void}; const f=a.m;",
		"const f=Math.floor;const {parseInt}=Number;const {map}=Array.prototype;class A extends Array<number>{};const a=new A;const m=a.map;",
		"const Math={floor(){}};const f=Math.floor;",
	}
}

// The strict-config limit is explicit. The six-rule runner still supports its
// previous configs; this expanded runner has not ported implicit-this messages.
func TestVolumeConfigGuardAndMutant(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShards(t)
}
