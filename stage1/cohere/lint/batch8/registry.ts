import type { Context } from './context.ts';
import { noOctalEscape } from './no_octal_escape.ts';
import { noUnexpectedMultiline } from './no_unexpected_multiline.ts';
import { noUnusedPrivateClassMembers } from './no_unused_private_class_members.ts';
import { noUselessConstructor } from './no_useless_constructor.ts';
import { preferTemplate } from './prefer_template.ts';
import { forwardRefUsesRef } from './forward_ref_uses_ref.ts';
import { jsxNoCommentTextnodes } from './jsx_no_comment_textnodes.ts';
import { noFindDomNode } from './no_find_dom_node.ts';
import { noIsMounted } from './no_is_mounted.ts';
import { noRedundantShouldComponentUpdate } from './no_redundant_should_component_update.ts';
export function visit(context: Context, index: number): void {
    noOctalEscape(context, index);
    noUnexpectedMultiline(context, index);
    noUnusedPrivateClassMembers(context, index);
    noUselessConstructor(context, index);
    preferTemplate(context, index);
    forwardRefUsesRef(context, index);
    jsxNoCommentTextnodes(context, index);
    noFindDomNode(context, index);
    noIsMounted(context, index);
    noRedundantShouldComponentUpdate(context, index);
    for(const child of context.node(index).children) {
        visit(context, child);
    }
}
