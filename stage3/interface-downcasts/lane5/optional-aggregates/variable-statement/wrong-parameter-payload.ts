// Original tsc declaration/read witness; adjacent helpers and carriers reduced.
interface ModifierLike {readonly value:number;}
interface VariableDeclaration {readonly value:number;}
interface VariableDeclarationList {readonly value:number;}
interface VariableStatement {readonly value:number;}
interface Base {readonly createVariableStatement:unknown;readonly updateVariableDeclarationList:(node:VariableDeclarationList,declarations:readonly VariableDeclaration[])=>VariableDeclarationList;}
interface Target {createVariableStatement(modifiers: readonly ModifierLike[] | undefined, declarationList: VariableDeclarationList | readonly VariableDeclaration[]): VariableStatement;readonly updateVariableDeclarationList:(node:VariableDeclarationList,declarations:readonly VariableDeclaration[])=>VariableDeclarationList;}
function inspect(node:VariableStatement):void {console.log(`${node.value}`);}
function probe(value:Base):void {const factory=value as Target;const node:VariableDeclarationList={value:5};const updatedDeclaration:VariableDeclaration={value:5};inspect(factory.createVariableStatement([{value:4}], factory.updateVariableDeclarationList(node, [updatedDeclaration])));}
probe({createVariableStatement:(modifiers:readonly {readonly value:string}[]|undefined,declarationList:VariableDeclarationList|readonly VariableDeclaration[]):VariableStatement=>({value:5+ +(modifiers===undefined?"0":modifiers[0]!.value)}),updateVariableDeclarationList:(node:VariableDeclarationList,declarations:readonly VariableDeclaration[]):VariableDeclarationList=>node});
