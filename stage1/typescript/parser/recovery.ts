// List terminators and diagnostics follow typescript-go's parsing contexts.
import { panic } from 'adamic';
import { tokenSpelling } from './grammar.ts';

interface ListDiagnosticInterface {
    readonly code: number;
    readonly message: string;
}

export function listTerminator(kind: string, context: string, lineBreak: boolean): boolean {
    if(kind === 'EndOfFile') {
        return true;
    }
    switch(context) {
        case 'source':
            return false;
        case 'block':
        case 'members':
        case 'object':
        case 'bindingObject':
        case 'specifiers':
        case 'attributes':
        case 'class':
        case 'enum':
        case 'switch':
            return kind === 'CloseBraceToken';
        case 'switchStatements':
            return ['CloseBraceToken', 'CaseKeyword', 'DefaultKeyword'].includes(kind);
        case 'parameters':
            return kind === 'CloseParenToken' || kind === 'CloseBracketToken';
        case 'arguments':
            return kind === 'CloseParenToken' || kind === 'SemicolonToken';
        case 'array':
        case 'bindingArray':
        case 'tuple':
            return kind === 'CloseBracketToken';
        case 'typeArguments':
            return kind !== 'CommaToken';
        case 'typeParameters':
            return [
                'GreaterThanToken',
                'OpenParenToken',
                'OpenBraceToken',
                'ExtendsKeyword',
                'ImplementsKeyword',
            ].includes(kind);
        case 'variables':
            return (
                ['SemicolonToken', 'CloseBraceToken', 'InKeyword', 'OfKeyword', 'EqualsGreaterThanToken'].includes(
                    kind,
                ) || lineBreak
            );
        default:
            return panic('unsupported parser list terminator');
    }
}

export function listDiagnostic(kind: string, context: string): ListDiagnosticInterface {
    switch(context) {
        case 'source':
            if(kind === 'DefaultKeyword') {
                return { code: 1005, message: "'export' expected." };
            }
            return { code: 1128, message: 'Declaration or statement expected.' };
        case 'block':
            return { code: 1128, message: 'Declaration or statement expected.' };
        case 'members':
            return { code: 1131, message: 'Property or signature expected.' };
        case 'arguments':
            return { code: 1135, message: 'Argument expression expected.' };
        case 'array':
            return { code: 1137, message: 'Expression or comma expected.' };
        case 'object':
            return { code: 1136, message: 'Property assignment expected.' };
        case 'bindingObject':
            return { code: 1180, message: 'Property destructuring pattern expected.' };
        case 'bindingArray':
            return { code: 1181, message: 'Array element destructuring pattern expected.' };
        case 'parameters':
            if(kind.endsWith('Keyword')) {
                return { code: 1390, message: `'${tokenSpelling(kind)}' is not allowed as a parameter name.` };
            }

            return { code: 1138, message: 'Parameter declaration expected.' };

        case 'variables':
            if(kind.endsWith('Keyword')) {
                return {
                    code: 1389,
                    message: `'${tokenSpelling(kind)}' is not allowed as a variable declaration name.`,
                };
            }

            return { code: 1134, message: 'Variable declaration expected.' };

        case 'typeParameters':
            return { code: 1139, message: 'Type parameter declaration expected.' };
        case 'typeArguments':
            return { code: 1140, message: 'Type argument expected.' };
        case 'tuple':
            return { code: 1110, message: 'Type expected.' };
        case 'specifiers':
            if(kind === 'FromKeyword') {
                return { code: 1005, message: "'}' expected." };
            }

            return { code: 1003, message: 'Identifier expected.' };

        case 'attributes':
            return { code: 1478, message: 'Identifier or string literal expected.' };
        case 'class':
            return {
                code: 1068,
                message: 'Unexpected token. A constructor, method, accessor, or property was expected.',
            };
        case 'enum':
            return { code: 1132, message: 'Enum member expected.' };
        case 'switch':
            return { code: 1130, message: "'case' or 'default' expected." };
        case 'switchStatements':
            return { code: 1129, message: 'Statement expected.' };
        default:
            return panic('unsupported parser list diagnostic');
    }
}
