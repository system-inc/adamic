'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const relative = 'src/compiler/utilities.ts';
const ownerName = 'createModuleNotFoundChain';

// Exact function text at ac44317c after every baseline adapter.
const before = [
    "export function createModuleNotFoundChain(sourceFile: SourceFile, host: TypeCheckerHost, moduleReference: string, mode: ResolutionMode, packageName: string): DiagnosticMessageChain {",
    "    const alternateResult = host.getResolvedModule(sourceFile, moduleReference, mode)?.alternateResult;",
    "    const alternateResultMessage = alternateResult && (getEmitModuleResolutionKind(host.getCompilerOptions()) === ModuleResolutionKind.Node10",
    "        ? [Diagnostics.There_are_types_at_0_but_this_result_could_not_be_resolved_under_your_current_moduleResolution_setting_Consider_updating_to_node16_nodenext_or_bundler, [alternateResult]] as const",
    "        : [",
    "            Diagnostics.There_are_types_at_0_but_this_result_could_not_be_resolved_when_respecting_package_json_exports_The_1_library_may_need_to_update_its_package_json_or_typings,",
    "            [alternateResult, alternateResult.includes(nodeModulesPathPart + \"@types/\") ? `@types/${mangleScopedPackageName(packageName)}` : packageName],",
    "        ] as const);",
    "    const result = alternateResultMessage",
    "        ? chainDiagnosticMessages(",
    "            /*details*/ undefined,",
    "            alternateResultMessage[0],",
    "            ...alternateResultMessage[1],",
    "        )",
    "        : host.typesPackageExists(packageName)",
    "        ? chainDiagnosticMessages(",
    "            /*details*/ undefined,",
    "            Diagnostics.If_the_0_package_actually_exposes_this_module_consider_sending_a_pull_request_to_amend_https_Colon_Slash_Slashgithub_com_SlashDefinitelyTyped_SlashDefinitelyTyped_Slashtree_Slashmaster_Slashtypes_Slash_1,",
    "            packageName,",
    "            mangleScopedPackageName(packageName),",
    "        )",
    "        : host.packageBundlesTypes(packageName)",
    "        ? chainDiagnosticMessages(",
    "            /*details*/ undefined,",
    "            Diagnostics.If_the_0_package_actually_exposes_this_module_try_adding_a_new_declaration_d_ts_file_containing_declare_module_1,",
    "            packageName,",
    "            moduleReference,",
    "        )",
    "        : chainDiagnosticMessages(",
    "            /*details*/ undefined,",
    "            Diagnostics.Try_npm_i_save_dev_types_Slash_1_if_it_exists_or_add_a_new_declaration_d_ts_file_containing_declare_module_0,",
    "            moduleReference,",
    "            mangleScopedPackageName(packageName),",
    "        );",
    "    if (result) result.repopulateInfo = () => ({ moduleReference, mode, packageName: packageName === moduleReference ? undefined : packageName });",
    "    return result;",
    "}",
].join('\r\n');

const first = before.indexOf('    const alternateResultMessage =');
const last = before.indexOf('        : host.typesPackageExists(packageName)');
const directCalls = [
    '    const result = alternateResult',
    '        ? getEmitModuleResolutionKind(host.getCompilerOptions()) === ModuleResolutionKind.Node10',
    '            ? chainDiagnosticMessages(',
    '                /*details*/ undefined,',
    '                Diagnostics.There_are_types_at_0_but_this_result_could_not_be_resolved_under_your_current_moduleResolution_setting_Consider_updating_to_node16_nodenext_or_bundler,',
    '                alternateResult,',
    '            )',
    '            : chainDiagnosticMessages(',
    '                /*details*/ undefined,',
    '                Diagnostics.There_are_types_at_0_but_this_result_could_not_be_resolved_when_respecting_package_json_exports_The_1_library_may_need_to_update_its_package_json_or_typings,',
    '                alternateResult,',
    '                alternateResult.includes(nodeModulesPathPart + "@types/") ? `@types/${mangleScopedPackageName(packageName)}` : packageName,',
    '            )',
    '',
].join('\r\n');
const after = before.slice(0, first) + directCalls + before.slice(last);

function findOwner(text) {
    assert.equal(ts.version, '6.0.3');
    const source = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'utilities.ts parse error');
    const owners = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === ownerName);
    assert.equal(owners.length, 1, `expected one ${ownerName}`);
    return owners[0];
}

function plan(text) {
    const owner = findOwner(text);
    const actual = owner.getText();
    if (actual === after) return text;
    assert.equal(actual, before, `unreviewed ${ownerName} snapshot`);
    return text.slice(0, owner.getStart()) + after + text.slice(owner.end);
}

module.exports = { relative, findOwner, before, after, plan };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    const file = path.join(process.argv[2], relative);
    const text = fs.readFileSync(file, 'utf8');
    const edited = plan(text);
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'source changed during adaptation');
    if (edited !== text) fs.writeFileSync(file, edited);
    console.log(`${ownerName}: ${edited === text ? 0 : 1} edit`);
}
