interface AutoGenerateInfo { readonly id: number }
interface EmitNode { autoGenerate: AutoGenerateInfo | undefined }
interface Identifier { emitNode: EmitNode | undefined }
interface GeneratedIdentifier { readonly emitNode: EmitNode & { autoGenerate: AutoGenerateInfo } }
const good: EmitNode & { autoGenerate: AutoGenerateInfo } = { autoGenerate: { id: 1 } };
const bad: EmitNode = { autoGenerate: undefined };
function createTempVariable(generated: true): GeneratedIdentifier;
function createTempVariable(generated: boolean): Identifier;
function createTempVariable(generated: boolean): Identifier { return { emitNode: generated ? good : bad }; }
const value = createTempVariable(true);
console.log((value.emitNode.autoGenerate === undefined).toString());
