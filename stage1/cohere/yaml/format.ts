import { CSTParser } from './cstParser.ts';
import { Composer } from './composer.ts';
import { UnistContext } from './unistContext.ts';
import { YAMLPrinter } from './printer.ts';
import { FormatResult } from './formatResult.ts';

// File-format defaults are Go cohere's formatoptions.PrettierDefaults: width 80, indentation 2, double quotes.
export function format(source: string): FormatResult {
    const result = new FormatResult();
    const bom = source.startsWith('\ufeff');
    const text = (bom ? source.slice(1) : source).split('\r\n').join('\n').split('\r').join('\n');
    if(text.trim() === '') {
        result.text = bom ? '\ufeff' : '';
        return result;
    }
    const parser = new CSTParser();
    parser.parse(text);
    const composer = new Composer(parser);
    composer.compose(text.length);
    for(const document of composer.documents) {
        const error = document.diagnostics.errors[0];
        if(error !== undefined) {
            result.kind = 'Error';
            result.message = error.message;
            return result;
        }
    }
    const context = new UnistContext(text, parser, composer);
    const root = context.parse();
    if(context.errorName !== '') {
        result.kind = 'Error';
        result.message = `${context.errorName}: ${context.errorMessage}`;
        return result;
    }
    const printer = new YAMLPrinter(context);
    result.text = (bom ? '\ufeff' : '') + printer.format(root);
    if(context.errorName !== '') {
        result.kind = 'Error';
        result.message = context.errorMessage;
    }
    return result;
}
