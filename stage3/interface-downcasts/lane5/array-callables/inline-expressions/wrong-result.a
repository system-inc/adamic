// Original tsc member and read witness; adjacent helpers/carriers reduced.
interface Expression {readonly kind:"expression";readonly value:number;}
interface TypeNode {readonly kind:"type";readonly value:number;}
interface CallExpression {readonly kind:"call";readonly value:number;readonly expression:Expression;}
interface Base {readonly inlineExpressions:unknown;readonly createOmittedExpression:()=>Expression;}
interface Target {inlineExpressions(expressions: readonly Expression[]): Expression;createOmittedExpression():Expression;}
function inspect(node:Expression):void {console.log(`${node.value}`);}
function probe(value:Base):void {const context:{readonly factory:Target}={factory:value as Target};const expressions:readonly Expression[]|undefined=[{kind:"expression",value:5}];inspect(context.factory.inlineExpressions(expressions!));}
probe({inlineExpressions:(expressions:readonly Expression[]):number=>7,createOmittedExpression:():Expression=>({kind:"expression",value:0})});
