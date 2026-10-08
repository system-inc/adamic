// Original tsc binding-read witness; adjacent carriers reduced.
type NodeFlags = number;
interface VariableDeclaration {readonly value:number;}
interface VariableDeclarationList {readonly value:number;}
interface Base {readonly createVariableDeclarationList:unknown;}
interface Target {createVariableDeclarationList(declarations: readonly VariableDeclaration[], flags?: NodeFlags): VariableDeclarationList;}
function inspect(node:VariableDeclarationList):void {console.log(`${node.value}`);}
function probe(value:Base):void {const factory=value as Target; const {createVariableDeclarationList: factoryCreateVariableDeclarationList} = factory; inspect(factoryCreateVariableDeclarationList([{value:5}], 4));}
probe({createVariableDeclarationList:(declarations:readonly VariableDeclaration[], flags?:NodeFlags):VariableDeclarationList=>({value:declarations[0]!.value+(flags ?? 0)})});
