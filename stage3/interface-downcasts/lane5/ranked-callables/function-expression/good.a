// Complete original declaration/read; adjacent carriers reduced.
interface Modifier {readonly value:number;}
interface AsteriskToken {readonly value:number;}
interface Identifier {readonly value:number;}
interface TypeParameterDeclaration {readonly value:number;}
interface ParameterDeclaration {readonly value:number;}
interface TypeNode {readonly value:number;}
interface Block {readonly value:number;}
interface FunctionExpression {readonly value:number;}
interface Base {readonly createFunctionExpression:unknown;}
interface Target {createFunctionExpression(modifiers: readonly Modifier[] | undefined, asteriskToken: AsteriskToken | undefined, name: string | Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[] | undefined, type: TypeNode | undefined, body: Block): FunctionExpression;}
function probe(value:Base):void {const factory=value as Target;const body:Block={value:3};console.log(`${factory.createFunctionExpression(undefined,undefined,undefined,undefined,undefined,undefined,body).value}`);}
probe({createFunctionExpression:(modifiers:readonly Modifier[]|undefined,asteriskToken:AsteriskToken|undefined,name:string|Identifier|undefined,typeParameters:readonly TypeParameterDeclaration[]|undefined,parameters:readonly ParameterDeclaration[]|undefined,type:TypeNode|undefined,body:Block):FunctionExpression=>({value:body.value+(parameters===undefined?0:parameters[0]!.value)})});
