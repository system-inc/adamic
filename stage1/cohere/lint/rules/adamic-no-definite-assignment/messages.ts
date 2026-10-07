import { PolicyMessage } from '../../helpers/policy_message.ts';
// Pinned Go-resolved catalog entry from helpers/testdata/catalog.json.
const catalog = new PolicyMessage("{\"adamic/no-definite-assignment\":{\"definiteAssignment\":{\"Text\":\"`!` here tells tsc this is assigned before it is read, and nothing checks it: read early, it is undefined while its type says otherwise. Assign it where it is declared or in the constructor, or type it with `| undefined` and handle the missing case.\",\"Options\":null}}}");
export const message = catalog.render('adamic/no-definite-assignment', 'definiteAssignment', new Map<string, string>(), []);
