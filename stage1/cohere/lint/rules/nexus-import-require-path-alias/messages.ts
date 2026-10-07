import { PolicyMessage } from '../../helpers/policy_message.ts';
const catalog = new PolicyMessage("{\"nexus/import-require-path-alias\": {\"useAlias\": {\"Text\": \"'{{importPath}}' climbs {{levels}} levels, which says how far up to walk rather than where it lands. A reader has to count directories to learn what it means. Import it as '{{suggestion}}', which reads the same from any file and survives this one moving.\", \"Options\": null}, \"useAliasInStrictRoot\": {\"Text\": \"'{{importPath}}' is relative, and everything under '{{root}}' is imported by alias so one path means one thing in every project that reads it. Import it as '{{suggestion}}'.\", \"Options\": null}}}");
export function message(id: string, path: string, levels: number, root: string, suggestion: string): string {
    const values = new Map<string, string>(); values.set('importPath', path); values.set('suggestion', suggestion);
    if(id === 'useAlias') { values.set('levels', String(levels)); } else { values.set('root', root); }
    return catalog.render('nexus/import-require-path-alias', id, values, []);
}
