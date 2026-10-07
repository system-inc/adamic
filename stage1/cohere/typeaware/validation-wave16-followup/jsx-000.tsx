declare namespace JSX {interface Element{} interface IntrinsicElements{[name:string]:unknown}} declare function Badge(p:{visible?:unknown}):JSX.Element;
declare const count:number, optional:number|undefined,big:bigint,flag:boolean,label:string; const view=<p>{count && <Badge/>}{flag && optional && 'x'}{(flag ? big : false) && 'x'}{label || count && 'x'}{(count || flag) && 'x'}</p>;
export {};
