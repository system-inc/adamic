// Static Go rule messages, not computed verdicts.
export function oneVarMessage(id: string, kind: string): string {
    switch(id) {
        case 'combineUninitialized':
            return `Combine this with the previous '${kind}' statement with uninitialized variables. This declares uninitialized variables in a second statement where the scope already has one. Reading the declarations of a scope means finding every statement that declares, and one list per scope makes that a single place to look.`;
        case 'combineInitialized':
            return `Combine this with the previous '${kind}' statement with initialized variables. This declares initialized variables in a second statement where the scope already has one. Reading the declarations of a scope means finding every statement that declares, and one list per scope makes that a single place to look.`;
        case 'splitUninitialized':
            return `Split uninitialized '${kind}' declarations into multiple statements. This declares several uninitialized variables in one statement. One binding per statement makes each declaration its own line to move, comment or delete, and keeps a diff that touches one variable from touching the others.`;
        case 'splitInitialized':
            return `Split initialized '${kind}' declarations into multiple statements. This declares several initialized variables in one statement. One binding per statement makes each declaration its own line to move, comment or delete, and keeps a diff that touches one variable from touching the others.`;
        case 'splitRequires':
            return `Split requires to be separated into a single block. This mixes require calls with other initializers in one declaration. Separating the requires keeps the module's dependencies readable as a block rather than interleaved with ordinary locals.`;
        case 'combine':
            return `Combine this with the previous '${kind}' statement. This declares variables in a second statement where the scope already has one. Reading the declarations of a scope means finding every statement that declares, and one list per scope makes that a single place to look.`;
        case 'split':
            return `Split '${kind}' declarations into multiple statements. This declares several variables in one statement. One binding per statement makes each declaration its own line to move, comment or delete, and keeps a diff that touches one variable from touching the others.`;
        default:
            return '';
    }
}
