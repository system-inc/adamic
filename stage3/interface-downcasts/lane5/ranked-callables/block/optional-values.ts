interface Statement {readonly value:number;}
interface Block {readonly value:number;}
interface Base {readonly createBlock:unknown;}
interface Target {createBlock(statements: readonly Statement[], multiLine?: boolean): Block;}
function probe(value:Base):void {const factory=value as Target;const returnStatement:Statement={value:3};const multiLine=true;const items:readonly Statement[]=[{value:3}];console.log(`${factory.createBlock(items,undefined).value}`);console.log(`${factory.createBlock(items,false).value}`);console.log(`${factory.createBlock(items,true).value}`);}
probe({createBlock:(statements:readonly Statement[],multiLine?:boolean):Block=>({value:statements[0]!.value+(multiLine===true?1:0)})});
