import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { OptionsJson } from '../../helpers/options_json.ts';
import { specifier } from '../nexus-import-no-forbidden-source/implementation.ts';
import { Alias, ordered, relative, normalizePath, normalizeDirectory, resolvePath, contains, aliasFor, quotedLike } from './paths.ts';
import { message } from './messages.ts';
export class Rule {
    readonly context: RuleContext; readonly filename: string; readonly root: string; readonly aliases: readonly Alias[]; readonly strict: string[] = [];
    constructor(context: RuleContext, filename: string, compilerAliases: readonly Alias[] | undefined, rawOptions: string = '') {
        this.context = context; this.filename = filename.split('\\').join('/');
        if(context.enabled('nexus/import-require-path-alias') && rawOptions === '' && context.settings.values.size > 0) { panic('NotYet: shared factory must supply raw alias options'); }
        const options = new OptionsJson(!context.enabled('nexus/import-require-path-alias') || rawOptions === '' ? 'null' : rawOptions); const object = options.parse();
        const configured: Alias[] = [];
        const root = options.field(object, 'repositoryRoot');
        let rootText = root < 0 ? '' : options.node(root).text;
        rootText = rootText.split('\\').join('/');
        this.root = rootText.endsWith('/') ? rootText.slice(0, -1) : rootText;
        const aliases = options.field(object, 'aliases');
        if(aliases >= 0) { for(const alias of options.node(aliases).children) { const directory = options.field(alias, 'directory'); const name = options.field(alias, 'alias'); configured.push(new Alias(directory < 0 ? '' : options.node(directory).text, name < 0 ? '' : options.node(name).text)); } }
        const derive = options.field(object, 'aliasesFromTsconfigPaths');
        if(derive >= 0 && options.node(derive).text === 'true') {
            if(compilerAliases === undefined) { panic('NotYet: shared compiler paths provider'); }
            for(const alias of compilerAliases) { configured.push(alias); }
        }
        this.aliases = ordered(configured);
        const strict = options.field(object, 'strictRoots');
        if(strict >= 0) { for(const item of options.node(strict).children) { this.strict.push(normalizeDirectory(options.node(item).text)); } }
    }
    visit(index: number): void {
        if(this.root === '' || this.aliases.length === 0) { return; }
        const importing = relative(this.root, this.filename);
        if(importing === undefined) { return; }
        const target = specifier(this.context, index);
        if(target < 0) { return; }
        const path = this.context.node(target).text;
        if(!path.startsWith('.')) { return; }
        const resolved = resolvePath(this.filename.slice(0, this.filename.lastIndexOf('/')), path);
        const local = relative(this.root, resolved);
        if(local === undefined || normalizePath(local).split('/').includes('internal')) { return; }
        const suggestion = aliasFor(this.aliases, local);
        if(suggestion === undefined) { return; }
        let id = ''; let strictRoot = '';
        for(const root of this.strict) { if(contains(root, importing)) { id = 'useAliasInStrictRoot'; strictRoot = root; break; } }
        const levels = path.split('../').length - 1;
        if(id === '' && levels >= 2) { id = 'useAlias'; }
        if(id === '') { return; }
        const original = this.context.source.slice(this.context.start(target), this.context.node(target).end);
        this.context.report(target, 'nexus/import-require-path-alias', id, message(id, path, levels, strictRoot, suggestion), 'fix', quotedLike(original, suggestion), '');
    }
}
