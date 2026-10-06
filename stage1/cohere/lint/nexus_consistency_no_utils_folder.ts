// Port of pinned Go cohere's nexus/consistency-no-utils-folder.
import type { RuleContext } from './rule_context.ts';
import { text, kind, range } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('nexus/consistency-no-utils-folder')) return;
    const rule = 'nexus/consistency-no-utils-folder';
    if(kind(ctx, index) !== 'SourceFile') return;
    for(const folder of ['_utils', 'utils'])
        if(ctx.parser.path.includes(`/${folder}/`) || ctx.parser.path.includes(`\\${folder}\\`))
            range(
                ctx,
                rule,
                folder === '_utils' ? 'noUnderscoreUtils' : 'noUtils',
                text(`nexus_consistency_no_utils_folder:${folder === '_utils' ? 'noUnderscoreUtils' : 'noUtils'}`),
                0,
                0,
            );
}
