// Scanner diagnostic text follows the typescript-go English messages.
import { panic } from 'adamic';

export function lexicalMessage(code: number): string {
    switch(code) {
        case 1002:
            return 'Unterminated string literal.';
        case 1010:
            return "'*/' expected.";
        case 1124:
            return 'Digit expected.';
        case 1125:
            return 'Hexadecimal digit expected.';
        case 1126:
            return 'Unexpected end of text.';
        case 1127:
            return 'Invalid character.';
        case 1160:
            return 'Unterminated template literal.';
        case 1161:
            return 'Unterminated regular expression literal.';
        case 1177:
            return 'Binary digit expected.';
        case 1178:
            return 'Octal digit expected.';
        case 1185:
            return 'Merge conflict marker encountered.';
        case 1198:
            return 'An extended Unicode escape value must be between 0x0 and 0x10FFFF inclusive.';
        case 1199:
            return 'Unterminated Unicode escape sequence.';
        case 1351:
            return 'An identifier or keyword cannot immediately follow a numeric literal.';
        case 1352:
            return 'A bigint literal cannot use exponential notation.';
        case 1353:
            return 'A bigint literal must be an integer.';
        case 1381:
            return "Unexpected token. Did you mean `{'}'}` or `&rbrace;`?";
        case 1382:
            return "Unexpected token. Did you mean `{'>'}` or `&gt;`?";
        case 1489:
            return 'Decimals with leading zeros are not allowed.';
        case 1490:
            return 'File appears to be binary.';
        case 6188:
            return 'Numeric separators are not allowed here.';
        case 6189:
            return 'Multiple consecutive numeric separators are not permitted.';
        case 18026:
            return "'#!' can only be used at the start of a file.";
        default:
            return panic('scanner diagnostic needs its message arguments');
    }
}
