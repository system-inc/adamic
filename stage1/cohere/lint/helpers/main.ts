import { panic, programArguments, readTextFile } from 'adamic';
import { OptionsJson } from './options_json.ts';
import { StrictOptions } from './strict_options.ts';
import { OptionSchema } from './option_schema.ts';
import { PolicyMessage, type MessageChoice } from './policy_message.ts';
function read(path: string): string { const value = readTextFile(path); if(value.kind === 'Error') { panic(value.message); } return value.text; }
const args = programArguments();
const cases = new OptionsJson(read(args[0] ?? panic('cases path'))); const root = cases.parse();
const catalog = new PolicyMessage(read(args[1] ?? panic('catalog path')));
function text(index: number, name: string): string { const value = cases.field(index, name); return value < 0 ? '' : cases.node(value).text; }
const definitions = cases.field(root, 'Definitions');
function schemaText(index: number): string { const key = text(index, 'Schema'); return cases.node(cases.field(definitions, key)).text; }
const schemas = new Map<string, OptionSchema>();
const targets = new Map<string, StrictOptions>();
for(const index of cases.node(cases.field(root, 'Cases')).children) {
    const kind = text(index, 'Kind'); const input = text(index, 'Input');
    if(kind === 'json') { const reader = new OptionsJson(input); console.log(reader.parse() >= 0 ? 'valid' : 'invalid'); }
    else if(kind === 'decode') {
        const key = text(index, 'Schema'); let target = targets.get(key);
        if(target === undefined) { target = new StrictOptions(schemaText(index)); targets.set(key, target); }
        console.log(target.check(input));
    }
    else if(kind === 'schema') {
        const key = text(index, 'Schema'); let schema = schemas.get(key);
        if(schema === undefined) { schema = new OptionSchema(schemaText(index)); schemas.set(key, schema); }
        console.log(schema.check(input));
    }
    else {
        const values = new Map<string, string>(); const raw = cases.field(index, 'Values');
        if(raw >= 0) { const value = cases.node(raw); for(let i = 0; i < value.keys.length; i++) { values.set(value.keys[i] ?? panic('value name'), cases.node(value.children[i] ?? panic('value')).text); } }
        const choices: MessageChoice[] = []; const choiceList = cases.field(index, 'Choices');
        if(choiceList >= 0) { for(const choice of cases.node(choiceList).children) { choices.push({rule:text(choice, 'Rule'), id:text(choice, 'Id'), phrase:text(choice, 'Phrase'), name:text(choice, 'Name')}); } }
        console.log(catalog.render(text(index, 'Rule'), text(index, 'Id'), values, choices));
    }
}
