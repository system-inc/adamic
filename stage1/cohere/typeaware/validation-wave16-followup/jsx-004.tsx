declare namespace JSX {interface Element{} interface IntrinsicElements{[name:string]:unknown}} declare function Badge(p:{visible?:unknown}):JSX.Element;
enum E{Off,On};enum F{One=1,Two=2};declare const e:E,f:F;const view=<p>{e&&'x'}{f&&'x'}</p>;
export {};
