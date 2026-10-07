"""Source witnesses for the unchanged production rule, including the stricter memo half."""
PREFIX='''export {};
declare const Ctx: {Provider:any};
declare const c:boolean;
declare const props:any;
declare const cache:any;
declare const store:any;
declare function opaque():any;
declare function createContext<T>(value:T):any;
declare function useMemo<T>(factory:()=>T,dependencies?:readonly unknown[]):T;
declare function useCallback<T>(factory:T,dependencies?:readonly unknown[]):T;
declare namespace React { class Component {} class PureComponent {} function createContext<T>(value:T):any; function memo<T>(value:T):T; function forwardRef<T>(value:T):T; }
'''
def controls():
 cases=[]
 for value in ['{}','[]','()=>{}','function(){}','class {}','new Object()','/x/u','<div/>','<></>','1','"x"','`x`','props.value','opaque()']:
  cases.append(f'function Component() {{ return <Ctx.Provider value={{{value}}}/>; }}')
 for value in ['{}','[]','()=>{}','function(){}','class {}','new Object()','/x/u','<div/>','<></>']:
  cases.append(f'function Component() {{ const value={value}; return <Ctx.Provider value={{value}}/>; }}')
 cases += [
  'function Component(){\n\n/* hi */ function value(){}\nreturn <Ctx.Provider value={value}/>;}',
  'const value={}; function Component(){return <Ctx.Provider value={value}/>;}',
  'function Outer(){const value={};return function Component(){return <Ctx.Provider value={value}/>;}}',
  'function Component(){const value={};if(c){return <Ctx.Provider value={value}/>;}}',
  'function Component(value:any){return <Ctx.Provider value={value}/>;}',
  'function Component(){const a=a;return <Ctx.Provider value={a}/>;}',
  'function Component(){const v={};return <Ctx.Provider value={v.x}/>;}',
  'function Component(){let v;return <Ctx.Provider value={v={}}/>;}',
  'function Component(){return <Ctx.Provider value={c?{}:[]}/>;}',
  'function Component(){return <Ctx.Provider value={c?props:[]}/>;}',
  'function Component(){return <Ctx.Provider value={props||{}}/>;}',
  'function Component(){return <Ctx.Provider value={({} as any)}/>;}',
  'function Component(){return <Ctx.Provider value={({} satisfies object)}/>;}',
  'function Component(){return <Ctx.Provider value={({})!}/>;}',
  'function component(){return <Ctx.Provider value={{}}/>;}',
  'function Component(){function helper(){return <Ctx.Provider value={{}}/>;}return helper();}',
  'function helper(){function Component(){return <Ctx.Provider value={{}}/>;}return Component;}',
  'const Component=()=> <Ctx.Provider value={{}}/>;',
  'const component=()=> <Ctx.Provider value={{}}/>;',
  'const thing=function Component(){return <Ctx.Provider value={{}}/>;};',
  'const Component=function thing(){return <Ctx.Provider value={{}}/>;};',
  'const Component=React.memo(()=> <Ctx.Provider value={{}}/>);',
  'const Component=React.forwardRef(()=> <Ctx.Provider value={{}}/>);',
  'class C extends React.Component { render(){return <Ctx.Provider value={{}}/>;} }',
  'class C extends React.PureComponent { other(){return <Ctx.Provider value={{}}/>;} }',
  'class C extends React.Component { f=()=> <Ctx.Provider value={{}}/>; }',
  'class C extends React.Component { f(){const helper=()=> <Ctx.Provider value={{}}/>;return helper();} }',
  'class C { Render(){return <Ctx.Provider value={{}}/>;} }',
  'class C extends Object { render(){return <Ctx.Provider value={{}}/>;} }',
  'const x=<Ctx.Provider value={{}}/>;',
  'function Émile(){return <Ctx.Provider value={{}}/>;}',
  'function Ω(){return <Ctx.Provider value={{}}/>;}',
  'function 𐐀(){return <Ctx.Provider value={{}}/>;}',
  'function Component(){const v\u200d={};return <Ctx.Provider value={v\u200d}/>;}',
  'const Other=createContext(0);function Component(){return <Other value={{}}/>;}',
  'const Other=(React.createContext(0));function Component(){return <Other value={{}}/>;}',
  'const Other=opaque();function Component(){return <Other value={{}}/>;}',
  'function Component(){return <Ctx.Other value={{}}/>;}',
  'function Component(){return <Ctx.Provider value/>;}',
  'function Component(){return <Ctx.Provider value="x"/>;}',
  'function Component(){return <Ctx.Provider value="x" value={{}}/>;}',
  'function Component(){return <Ctx.Provider {...props}/>;}',
  '/* 😀 é */\r\nfunction Component(){\r\n const value={};\r\n return <Ctx.Provider value={value}/>;\r\n}',
 ]
 for hook in ['useMemo(()=>({}))','useMemo(()=>({}),[])','useMemo(()=>cache)','useCallback(()=>{})','useCallback(()=>{},[])','React.useMemo(()=>({}))','useMemo(()=>({}),props.deps)']:
  cases.append(f'function Component(){{ const value={hook}; return <Ctx.Provider value={{value}}/>; }}')
 for dependency in ['{}','[]','()=>{}','/x/','new Object()','<div/>','<></>']:
  cases.append(f'function Component(){{const dep={dependency};const value=useMemo(()=>({{dep}}),[dep]);return <Ctx.Provider value={{value}}/>;}}')
 cases += [
  'function Component(){let dep={};const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function Component(){const {dep}=props;const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'const dep={};function Component(){const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function make(){return {};}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'async function make(){return 1;}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function* make(){yield 1;}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function make():number{return {} as any;}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function Component(){const dep=Math.random();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function make(){if(c)return cache;const value={};store.value=value;return value;}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function make(){if(c)return cache;const value={};value.touch();return value;}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function make(){if(c)return cache;const value={};const alias=value;return alias;}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function make(){if(c)return cache;const value={};opaque(value);return value;}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function make(){if(c)return cache;const value={};return value;}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function make(){return cache??{};}function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function Component(){const dep=useMemo(()=>({}));const value=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={value}/>;}',
  'function Component(){const dep={};const inner=useMemo(()=>({dep}),[dep]);const value=useMemo(()=>inner,[inner]);return <Ctx.Provider value={value}/>;}',
  'function Component(){const dep={};const value=useMemo(()=>cache,[dep]);return <Ctx.Provider value={value}/>;}',
  'function Component(){const dep={};const value=useMemo(()=>c?dep:cache,[dep]);return <Ctx.Provider value={value}/>;}',
  'function Component(){const value=(useMemo(()=>({})))!;return <Ctx.Provider value={value}/>;}',
  'function Component(){const value=useMemo(()=>({}),[{}]);return <Ctx.Provider value={value}/>;}',
  'function Component(){return <Ctx.Provider value={useMemo(()=>({}),[{}])}/>;}',
 ]
 for count in [1,3,8,9,16,17]:
  chain='const dep0={};'+''.join(f'const dep{i}=dep{i-1};' for i in range(1,count))
  cases.append(f'function Component(){{{chain}const value=useMemo(()=>({{}}),[dep{count-1}]);return <Ctx.Provider value={{value}}/>;}}')
 return cases
