interface Statement {readonly value:number;}
interface Block {readonly value:number;}
interface Base {readonly createBlock:unknown;}
interface Target {createBlock(statements: readonly Statement[], multiLine?: boolean): Block;}
function probe(value:Base):void {const factory=value as Target;const returnStatement:Statement={value:3};const multiLine=true;console.log(`${factory.createBlock([returnStatement],multiLine).value}`);}
probe({createBlock:(statements:readonly {readonly value:string}[],multiLine?:boolean):{readonly value:string|number}=>({value:statements[0]!.value})});
