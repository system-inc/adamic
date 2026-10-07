// Render a Go-validated, language-resolved policy catalog. Values are substituted in Go's sorted order.
import { panic } from 'adamic';
import { OptionsJson } from './options_json.ts';
export interface MessageChoice { readonly rule: string; readonly id: string; readonly phrase: string; readonly name: string; }
export class PolicyMessage {
    readonly catalog: OptionsJson;
    readonly root: number;
    constructor(text: string) {
        this.catalog = new OptionsJson(text); this.root = this.catalog.parse();
        if(this.root < 0) { panic('invalid generated message catalog'); }
    }
    template(rule: string, id: string): number {
        const rules = this.catalog.field(this.root, rule);
        const template = rules < 0 ? -1 : this.catalog.field(rules, id);
        if(template < 0) { panic(`policy/messages: ${rule} has no message "${id}"`); }
        return template;
    }
    names(text: string): string[] {
        const seen = new Set<string>(); let rest = text;
        while(true) {
            const start = rest.indexOf('{{'); const stray = rest.indexOf('}}');
            if(start < 0) { if(stray >= 0) { panic('malformed generated message placeholder'); } break; }
            if(stray >= 0 && stray < start) { panic('malformed generated message placeholder'); }
            const end = rest.indexOf('}}', start + 2);
            if(end < 0) { panic('malformed generated message placeholder'); }
            const name = rest.slice(start + 2, end);
            if(name === '' || name.charCodeAt(0) < 97 || name.charCodeAt(0) > 122 || ![...name].every(c => (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'))) { panic('malformed generated message placeholder'); }
            seen.add(name); rest = rest.slice(end + 2);
        }
        return [...seen].sort((a, b) => a < b ? -1 : a > b ? 1 : 0);
    }
    render(rule: string, id: string, values: ReadonlyMap<string, string>, choices: readonly MessageChoice[]): string {
        const template = this.template(rule, id);
        const raw = this.catalog.field(template, 'Text'); const options = this.catalog.field(template, 'Options');
        let text = this.catalog.node(raw).text;
        const chosen = new Set<string>();
        for(const choice of choices) {
            const phrase = options < 0 ? -1 : this.catalog.field(options, choice.phrase);
            const option = phrase < 0 ? -1 : this.catalog.field(phrase, choice.name);
            if(choice.rule !== rule || choice.id !== id || option < 0 || chosen.has(choice.phrase)) {
                panic(`policy/messages: ${rule} "${id}" was given the option "${choice.name}" of "${choice.phrase}", which it does not take`);
            }
            chosen.add(choice.phrase); text = text.split(`<<${choice.phrase}>>`).join(this.catalog.node(option).text);
        }
        const count = options < 0 ? 0 : this.catalog.node(options).keys.length;
        if(chosen.size !== count) { panic(`policy/messages: ${rule} "${id}" picks from ${count} phrases, and was given ${chosen.size} options`); }
        const needed = this.names(text);
        if(values.size !== needed.length) { panic(`policy/messages: ${rule} "${id}" uses the values [${needed.join(' ')}], and was given ${values.size}`); }
        for(const name of needed) {
            const value = values.get(name);
            if(value === undefined) { panic(`policy/messages: ${rule} "${id}" uses the value "${name}", and was not given it`); }
            text = text.split(`{{${name}}}`).join(value);
        }
        return text;
    }
}
