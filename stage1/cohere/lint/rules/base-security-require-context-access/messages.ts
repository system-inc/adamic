import { PolicyMessage } from '../../helpers/policy_message.ts';
const catalog = new PolicyMessage("{\"base/security-require-context-access\":{\"missingProtector\":{\"Text\":\"This injects `{{contextKey}}` from the request context but carries no decorator establishing access to it, so the value arrives whether or not anyone verified the request and any authorization decision made from it is made on unverified input. Add {{protectors}}, to this member or to its class.\",\"Options\":null}}}");
export function missingProtector(key: string, requires: readonly string[]): string {
    const values = new Map<string, string>();
    values.set('contextKey', key);
    values.set('protectors', requires.map(name => '`@' + name + '()`').join(' or '));
    return catalog.render('base/security-require-context-access', 'missingProtector', values, []);
}
