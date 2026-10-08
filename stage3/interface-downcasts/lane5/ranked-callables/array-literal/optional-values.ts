interface Expression {readonly value:number;}
interface ArrayLiteralExpression {readonly value:number;}
interface Base {readonly createArrayLiteralExpression:unknown;}
interface Target {createArrayLiteralExpression(elements?: readonly Expression[], multiLine?: boolean): ArrayLiteralExpression;}
function probe(value:Base):void {const factory=value as Target;const elements:readonly Expression[]=[{value:3}];const items:readonly Expression[]=[{value:3}];console.log(`${factory.createArrayLiteralExpression().value}`);console.log(`${factory.createArrayLiteralExpression(items,undefined).value}`);console.log(`${factory.createArrayLiteralExpression(items,false).value}`);console.log(`${factory.createArrayLiteralExpression(items,true).value}`);}
probe({createArrayLiteralExpression:(elements?:readonly Expression[],multiLine?:boolean):ArrayLiteralExpression=>({value:(elements===undefined?0:elements[0]!.value)+(multiLine===true?1:0)})});
