# JSON and config public declaration design

## Decision and provenance

This is a declaration-only design probe, not an adaptation. It inventories 16 public members and proposes 16 retyped declaration lines plus two published type definitions. The exact incremental API diff is 144 lines including headers and context: 20 added, 16 removed, net 4. The unchecked proposal does not compile. Runtime baseline results are recorded separately below.

Branch `typescript/json-design` starts at fetched main `c74051ffc5ee94f4b7f732b1532cb219032b183d`. Upstream is TypeScript 6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. The checker is stock 6.0.3 from the locked `stage3/api` installation. Setup exited 0 in 410.207 seconds; full `stage3/apply.sh` exited 0. The final control normal build and stock `tsc -b src --force` both exited 0. An earlier build overlapped apply and is superseded by the final build after apply completed. No suite overlapped setup, adaptation, or preparation.

Read before proposing: ANY5.md, all of CLASSIFICATION.md including its 38 JSON/config sites, adaptation 43's README, adapter, rules and proofs, adaptation 91's README, generate-api.cjs and sanctioned-api.json. The current generator reconstructs 223 sanctioned owners from pinned source and proof records, including 28 adaptation-40 lines. Its sanction ledger allows only `api/typescript.d.ts` to differ. Publishing definitions and changing these 16 lines would require a future explicit sanction review. This branch changes neither generator nor ledger.

## Published type and boundaries

Keep adaptation 43's names and definitions, removing their two `@internal` markers in the scratch proposal:

```ts
export type AdamicJsonRecoveryValue =
    string | number | boolean | null |
    AdamicJsonRecoveryValue[] | AdamicJsonRecoveryObject;
export interface AdamicJsonRecoveryObject {
    [key: string]: AdamicJsonRecoveryValue | undefined;
}
```

There is one recursive value domain, with an object interface for indexing it. Strict JSON values fit this domain. The tsconfig parser also retains invalid object property values as undefined, filters invalid array elements, and recovers an object root from malformed config input. An object index without undefined would misdescribe those existing results. `convertToObject` permits scalar, array and absent results; it must return the full value union plus undefined. The two config readers return optional recovery objects, not unknown, because their producer is the object-root tsconfig converter.

Use unknown at caller boundaries. The public `json` and `jsonOptions` inputs accept arbitrary JavaScript values. Both raw members can carry these values unchanged, and a caller can construct a ParsedCommandLine or ParsedTsconfig directly. Declaring them recovery objects would claim validation that does not happen. Protocol base arguments/bodies, plugin config, telemetry payload and the localization hook also admit caller values; specialized protocol interfaces retain their existing precise fields. Inbound server messages originate in JSON.parse, but Session.parseMessage currently asserts Request without checking its shape. The eventual implementation must check at that read and before command-specific use. This probe adds no such code.

For the later internal pipeline, the same recovery union can describe JSON.parse results and parser values; `tryParseJson` would additionally allow undefined on failure. Unknown must be checked before indexing or assigning a more specific contract. Keep the distinction between validating an object and validating each option, array element or package field. A cast to the desired map is not that check. Internal JSON.parse and helper signatures are outside this public-declaration-only measurement; their extra errors have not been measured here.

## Before and after

These are the exact changed public lines, stripped of indentation. Return-property rows identify the containing function. Full function signatures are retained. The API diff below supplies namespace context.

| Public member | Before | After |
| --- | --- | --- |
| `ts.server.protocol.Request.arguments` | `arguments?: any;` | `arguments?: unknown;` |
| `ts.server.protocol.Event.body` | `body?: any;` | `body?: unknown;` |
| `ts.server.protocol.Response.body` | `body?: any;` | `body?: unknown;` |
| `ts.server.protocol.ConfigurePluginRequestArguments.configuration` | `configuration: any;` | `configuration: unknown;` |
| `ts.server.protocol.TelemetryEventBody.payload` | `payload: any;` | `payload: unknown;` |
| `ts.server.PluginCreateInfo.config` | `config: any;` | `config: unknown;` |
| `ts.server.PluginModule.onConfigurationChanged` | `onConfigurationChanged?(config: any): void;` | `onConfigurationChanged?(config: unknown): void;` |
| `ts.ParsedCommandLine.raw` | `raw?: any;` | `raw?: unknown;` |
| `ts.readConfigFile.config` | `config?: any;` | `config?: AdamicJsonRecoveryObject;` |
| `ts.parseConfigFileTextToJson.config` | `config?: any;` | `config?: AdamicJsonRecoveryObject;` |
| `ts.convertToObject` | `function convertToObject(sourceFile: JsonSourceFile, errors: Diagnostic[]): any;` | `function convertToObject(sourceFile: JsonSourceFile, errors: Diagnostic[]): AdamicJsonRecoveryValue \| undefined;` |
| `ts.parseJsonConfigFileContent` | `function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine;` | `function parseJsonConfigFileContent(json: unknown, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine;` |
| `ts.convertCompilerOptionsFromJson` | `function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): {` | `function convertCompilerOptionsFromJson(jsonOptions: unknown, basePath: string, configFileName?: string): {` |
| `ts.convertTypeAcquisitionFromJson` | `function convertTypeAcquisitionFromJson(jsonOptions: any, basePath: string, configFileName?: string): {` | `function convertTypeAcquisitionFromJson(jsonOptions: unknown, basePath: string, configFileName?: string): {` |
| `ts.ParsedTsconfig.raw` | `raw: any;` | `raw: unknown;` |
| `ts.LanguageServiceHost.getLocalizedDiagnosticMessages` | `getLocalizedDiagnosticMessages?(): any;` | `getLocalizedDiagnosticMessages?(): unknown;` |

## Inventory, producers, readers and writers

[JSON-DESIGN.references.tsv](JSON-DESIGN.references.tsv) lists every measured symbol reference, contextual object initializer and raw/extendsRaw alias use inside src/compiler, src/services and src/server, with file, line, column and full whitespace-folded statement. It includes specialized protocol overrides and their references, since TypeScript findReferences groups the base and override symbols. Such entries do not all acquire errors. Bindings that findReferences labels writes can be reads of a member; the explicit alias-use rows retain their subsequent uses. Reference locations are in the final adapted control. Error locations below are in the proposal; removing two internal-marker lines shifts later commandLineParser.ts locations by two lines.

The inventory uses a stock language service over compiler, services, server and their source dependencies, resolves target declarations by owner and member, asks findReferences, identifies contextual object writes by checker symbol identity, and follows the two raw aliases by local symbol identity. The following table records producers and transfers that property-reference lookup alone does not express. Public callers are outside the in-repository reader/writer scope.

| Member | Origin and complete in-scope production/consumption summary |
| --- | --- |
| ParsedCommandLine.raw | Caller or tsconfig parser via ParsedTsconfig.raw. commandLineParser.ts:3080 destructures it and :3104 stores it. Reads: commandLineParser.ts:3355; watchPublic.ts:979; tsbuildPublic.ts:1245; server/project.ts:3157; server/editorServices.ts:2922-2925. No other direct writes in scope. |
| readConfigFile.config | Tsconfig parser, not JSON.parse. commandLineParser.ts:2269 forwards parseConfigFileTextToJson or constructs config {} on read failure. No direct member reader in the three requested directories. jsTyping consumers are measured separately below. |
| parseConfigFileTextToJson.config | Tsconfig parser. commandLineParser.ts:2280 writes convertConfigFileToObject's object result; :2269 forwards the complete result through readConfigFile. utilities.ts:7795 calls it and :7797 reads looseResult.config into an any local. |
| convertToObject return | Tsconfig JSON AST converter. commandLineParser.ts:2472-2474 returns convertToJson; :3425 consumes its result for the circularity raw result. No JSON.parse or arbitrary caller value is manufactured by this converter. |
| parseJsonConfigFileContent.json | Caller. commandLineParser.ts:3019 forwards the parameter to parseJsonConfigFileContentWorker. No in-scope caller, reassignment or direct property read of this parameter. Internal worker remains any. |
| convertCompilerOptionsFromJson.jsonOptions | Caller. commandLineParser.ts:3740 forwards to convertCompilerOptionsFromJsonWorker. No in-scope caller, reassignment or direct property read of this parameter. |
| convertTypeAcquisitionFromJson.jsonOptions | Caller. commandLineParser.ts:3746 forwards to convertTypeAcquisitionFromJsonWorker. No in-scope caller, reassignment or direct property read of this parameter. |
| ParsedTsconfig.raw | Caller path :3519; tsconfig object path :3600; circularity result :3425 uses caller json or convertToObject. Direct reads, inheritance writes and raw/extendsRaw aliases are all in commandLineParser.ts. Full locations and statements are in the attachment; inherited include/exclude/files and compileOnSave writes do not validate the original raw object. |
| protocol.Request.arguments | JSON.parse at server/session.ts:3948 returns an asserted Request, or a caller supplies a request. No direct base-member assignment in these directories. The attachment lists every dispatch read, specialized override and related reader, including editorServices.ts:1196. Two base Request implementation handlers at session.ts:3504 and :3507 lose their any escape; most handlers already have a specialized argument contract. |
| protocol.Event.body | Caller through generic event helpers. server/session.ts:479 constructs the base Event body in toEvent; Session.event at :1264-1265 forwards it to send. Specialized event body declarations and references are included in the attachment. Serialization sends the complete event. No base body field read in scope. |
| protocol.Response.body | Caller/result objects. server/session.ts:1289, :1296, :1300 and :1304 assign arrays, the metadata-stripped body or info. Specialized response declarations are listed in the attachment. No other base body writer or direct base field read in scope. |
| ConfigurePluginRequestArguments.configuration | Caller, or an inbound JSON.parse request. server/editorServices.ts:5581 forwards it to each project; :5586 stores the override. The source declaration and both reads are in the attachment. |
| TelemetryEventBody.payload | Caller/event data. server/session.ts:1181 constructs payload from event.data; typingInstallerAdapter.ts:180 constructs the specialized installed-typings payload. Specialized declaration/reference included in the attachment. No base payload field reader in scope. |
| PluginCreateInfo.config | Caller configEntry, often from tsconfig plugins or a protocol override; synthetic global plugin entries also exist. server/project.ts:2173 stores it in info, and :2182 passes info to pluginModule.create. No in-scope member read; plugins supplied by callers consume it. |
| PluginModule.onConfigurationChanged config parameter | Caller. server/project.ts:2202 checks the hook and :2203 calls it with configuration. The callback implementation is supplied by the plugin, so there is no in-scope parameter body. |
| LanguageServiceHost.getLocalizedDiagnosticMessages return | Caller hook. services/services.ts:1656 binds it and passes it to maybeSetLocalizedDiagnosticMessages. There is no in-scope implementation writing this member. Localized files also have a separate JSON.parse boundary in compiler/utilitiesPublic.ts:749; changing that internal boundary is not part of this probe. |

Excluded remaining public any declarations: timer handles and callback tuples; nominal marker slots; protocol's ChangePropertyTypes mapped substitution constraint; CommandNames' runtime command table; and CompressedData.data. CompressedData has no reader, writer or JSON producer in these source directories and describes opaque compressed host data. These are not a JSON/config-value contract. The scan examined the complete generated public declaration, including server and services sections, rather than treating the 38 compiler census tokens as 38 public members. Internal tryParseJson, convertToJson, JSON conversion notifier fields and config validators remain internal and are not counted as public members.

## Reader errors and checks needed

Stock 6.0.3 ran `tsc -p src/compiler`, services, server and harness with their inherited upstream options, adding only `--pretty false` for stable logs. No strictness option was relaxed. Declaration dependencies were refreshed in order: compiler, jsTyping, services, typingsInstallerCore, deprecatedCompat, server, harness, typescript. The control had zero diagnostics in every project. Default declaration emit still produced dependencies after proposal errors, so downstream projects saw the proposed types. The ten jsTyping errors are reported even though jsTyping was not one of the four requested projects. The separate upstream src/testRunner project was also checked after the runtime suites, to cover the full test build as well as src/harness; its control has zero diagnostics.

| Project | New errors | Exit |
| --- | ---: | ---: |
| compiler | 17 | 2 |
| services | 0 | 0 |
| server | 6 | 2 |
| harness | 5 | 2 |
| testRunner | 1 | 2 |
| jsTyping | 10 | 2 |
| typingsInstallerCore | 0 | 0 |
| deprecatedCompat | 0 | 0 |
| typescript | 0 | 0 |

[JSON-DESIGN.errors.tsv](JSON-DESIGN.errors.tsv) retains every new diagnostic, including code, exact file/line/column and complete message. There are 39 diagnostics in total, not 39 distinct checks. Counts by file:

| File | New diagnostics |
| --- | ---: |
| `src/compiler/commandLineParser.ts` | 17 |
| `src/harness/client.ts` | 3 |
| `src/harness/harnessLanguageService.ts` | 2 |
| `src/jsTyping/jsTyping.ts` | 10 |
| `src/server/editorServices.ts` | 4 |
| `src/server/session.ts` | 2 |
| `src/testRunner/unittests/tsserver/session.ts` | 1 |

The compiler errors require proving raw object presence before field access or hasProperty, then checking the option/inheritance values. The four editorServices errors require an object check before the telemetry property-presence reads. The two session errors require validating command arguments as FileLocationRequestArgs. Harness response reads require checking response bodies, while plugin fixtures require checking config fields. The testRunner echo handler at src/testRunner/unittests/tsserver/session.ts:525 forwards unknown request arguments into HandlerResponse, which permits only an object or undefined; that transfer also needs checking. jsTyping needs checked string maps and package/dependency object fields; its downstream array/key errors are consequences of those missing checks.

Zero services errors is not proof that the host result is validated: strictBindCallApply is false in upstream's options, so bind erases the new unknown return before maybeSetLocalizedDiagnosticMessages. Likewise, public JSON parameters flow straight into unchanged any-typed internal workers, and canJsonReportNoInputFiles still accepts any. These are recorded transfers that still need design work even though declaration-only edits do not diagnose them. Existing assertions, specialized interfaces, any slots and skipLibCheck also bound what this experiment can expose.

## Exact generated API diff

Compared with the final adapted control's generated public API, not with pristine TypeScript's already-sanctioned differences. The following is the complete 144-line unified diff. It is generated by upstream's dtsBundler from actual emitted declarations; no lane sanction or baseline is edited to accept it.

```diff
--- control/api/typescript.d.ts
+++ proposal/api/typescript.d.ts
@@ -149,7 +149,7 @@
                 /**
                  * Object containing arguments for the command
                  */
-                arguments?: any;
+                arguments?: unknown;
             }
             /**
              * Request to reload the project structure for all the opened files
@@ -169,7 +169,7 @@
                 /**
                  * Event-specific information
                  */
-                body?: any;
+                body?: unknown;
             }
             /**
              * Response by server to client request message.
@@ -196,7 +196,7 @@
                 /**
                  * Contains message body if success === true.
                  */
-                body?: any;
+                body?: unknown;
                 /**
                  * Contains extra information that plugin can include to be passed on
                  */
@@ -1199,7 +1199,7 @@
             }
             export interface ConfigurePluginRequestArguments {
                 pluginName: string;
-                configuration: any;
+                configuration: unknown;
             }
             export interface ConfigurePluginRequest extends Request {
                 command: CommandTypes.ConfigurePlugin;
@@ -2388,7 +2388,7 @@
             }
             export interface TelemetryEventBody {
                 telemetryEventName: string;
-                payload: any;
+                payload: unknown;
             }
             export type TypesInstallerInitializationFailedEventName = "typesInstallerInitializationFailed";
             export interface TypesInstallerInitializationFailedEvent extends Event {
@@ -2842,12 +2842,12 @@
             languageServiceHost: LanguageServiceHost;
             serverHost: ServerHost;
             session?: Session<unknown>;
-            config: any;
+            config: unknown;
         }
         interface PluginModule {
             create(createInfo: PluginCreateInfo): LanguageService;
             getExternalFiles?(proj: Project, updateLevel: ProgramUpdateLevel): string[];
-            onConfigurationChanged?(config: any): void;
+            onConfigurationChanged?(config: unknown): void;
         }
         interface PluginModuleWithName {
             name: string;
@@ -7241,7 +7241,7 @@
         fileNames: string[];
         projectReferences?: readonly ProjectReference[] | undefined;
         watchOptions?: WatchOptions | undefined;
-        raw?: any;
+        raw?: unknown;
         errors: Diagnostic[];
         wildcardDirectories?: MapLike<WatchDirectoryFlags>;
         compileOnSave?: boolean;
@@ -9237,7 +9237,7 @@
      * @param fileName The path to the config file
      */
     function readConfigFile(fileName: string, readFile: (path: string) => string | undefined): {
-        config?: any;
+        config?: AdamicJsonRecoveryObject;
         error?: Diagnostic | undefined;
     };
     /**
@@ -9246,7 +9246,7 @@
      * @param jsonText The text of the config file
      */
     function parseConfigFileTextToJson(fileName: string, jsonText: string): {
-        config?: any;
+        config?: AdamicJsonRecoveryObject;
         error?: Diagnostic | undefined;
     };
     /**
@@ -9257,7 +9257,7 @@
     /**
      * Convert the json syntax tree into the json value
      */
-    function convertToObject(sourceFile: JsonSourceFile, errors: Diagnostic[]): any;
+    function convertToObject(sourceFile: JsonSourceFile, errors: Diagnostic[]): AdamicJsonRecoveryValue | undefined;
     /**
      * Parse the contents of a config file (tsconfig.json).
      * @param json The contents of the config file to parse
@@ -9265,7 +9265,7 @@
      * @param basePath A root directory to resolve relative path entries in the config
      *    file to. e.g. outDir
      */
-    function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine;
+    function parseJsonConfigFileContent(json: unknown, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine;
     /**
      * Parse the contents of a config file (tsconfig.json).
      * @param jsonNode The contents of the config file to parse
@@ -9274,11 +9274,11 @@
      *    file to. e.g. outDir
      */
     function parseJsonSourceFileConfigFileContent(sourceFile: TsConfigSourceFile, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine;
-    function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): {
+    function convertCompilerOptionsFromJson(jsonOptions: unknown, basePath: string, configFileName?: string): {
         options: CompilerOptions;
         errors: Diagnostic[];
     };
-    function convertTypeAcquisitionFromJson(jsonOptions: any, basePath: string, configFileName?: string): {
+    function convertTypeAcquisitionFromJson(jsonOptions: unknown, basePath: string, configFileName?: string): {
         options: TypeAcquisition;
         errors: Diagnostic[];
     };
@@ -9305,8 +9305,12 @@
     interface ParseConfigFileHost extends ParseConfigHost, ConfigFileDiagnosticsReporter {
         getCurrentDirectory(): string;
     }
+    type AdamicJsonRecoveryValue = string | number | boolean | null | AdamicJsonRecoveryValue[] | AdamicJsonRecoveryObject;
+    interface AdamicJsonRecoveryObject {
+        [key: string]: AdamicJsonRecoveryValue | undefined;
+    }
     interface ParsedTsconfig {
-        raw: any;
+        raw: unknown;
         options?: CompilerOptions;
         watchOptions?: WatchOptions | undefined;
         typeAcquisition?: TypeAcquisition;
@@ -10082,7 +10086,7 @@
         getScriptVersion(fileName: string): string;
         getScriptSnapshot(fileName: string): IScriptSnapshot | undefined;
         getProjectReferences?(): readonly ProjectReference[] | undefined;
-        getLocalizedDiagnosticMessages?(): any;
+        getLocalizedDiagnosticMessages?(): unknown;
         getCancellationToken?(): HostCancellationToken;
         getCurrentDirectory(): string;
         getDefaultLibFileName(options: CompilerOptions): string;
```

## Suite and baseline movement

Both complete suites ran serially through `stage3/oracle/run.sh --prepared`, all runners, no test filter, four workers, a 1,800-second suite limit and a 1,400 MiB heap limit per Node process. Setup and all preparation had finished before the control started. The proposal uses separate declaration/runtime emission because its normal build rejects the readers; this is explicitly not a passing checked lane.

| Tree | Passing | Failing | Pending | Suite seconds | Baseline files |
| --- | ---: | ---: | ---: | ---: | --- |
| control | 106,366 | 1 | 0 | 382.858 | `api/typescript.d.ts` |
| probe | 106,365 | 2 | 0 | 381.251 | `APISample_parseConfig.errors.txt`, `api/typescript.d.ts` |

Each oracle exits 1. Control has only the existing public API acknowledgment failure. The proposal has that failure and one compiler API sample failure. Total moved baseline files: control **1**, proposal **2**. The complete list besides `api/typescript.d.ts` has **1** file:

- `APISample_parseConfig.errors.txt` (new diagnostic baseline, 56 lines).

The sample destructures the config result, then indexes `config.config["compilerOptions"]` at virtual APISample_parseConfig.ts:27:56 (tests/cases/compiler/APISample_parseConfig.ts:41 in the case input). The new type exposes three diagnostics there: TS18048 (config may be undefined), TS18049 (config.config may be null or undefined), and TS7053 (compilerOptions is not a valid key on every member of the recursive value union). These are suite fixture diagnostics, separate from the 39 project-check diagnostics. [JSON-DESIGN.baselines.json](JSON-DESIGN.baselines.json) retains the complete new baseline text, original line endings and its SHA-256. The sample is not edited. This is an actual public consumer compatibility effect even though emitted JavaScript is identical. No other baseline moves, and no failure is suppressed or accepted by changing a baseline. The local API baseline files are byte-identical to their respective generated public declaration artifacts. Their incremental difference is exactly the 144-line diff above. All ten JavaScript artifact hashes remain identical after both suites.

[JSON-DESIGN.measurements.json](JSON-DESIGN.measurements.json) retains the commands, exits, project/file counts, oracle summaries, declaration hashes and all ten JavaScript hashes. The prepared oracle's candidate build record distinguishes successful emission from failed typechecking. Actual wrapper commands were:

```sh
NODE_OPTIONS=--max-old-space-size=1400 bash stage3/oracle/run.sh /workspace/scratch/json-control /workspace/scratch/json-oracle-control --prepared --workers 4 --limit-seconds 1800 > /workspace/scratch/oracle-control.log 2>&1
NODE_OPTIONS=--max-old-space-size=1400 bash stage3/oracle/run.sh /workspace/scratch/json-probe /workspace/scratch/json-oracle-probe --prepared --workers 4 --limit-seconds 1800 > /workspace/scratch/oracle-probe.log 2>&1
```

## Reproduction and limits

Preparation logs are files under `/workspace/scratch`, never piped to head or tail. Required setup used exactly:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
```

The initial apply failed only because no product or pinned source checkout was available. After cloning v6.0.3 and installing stage3/api's locked packages, the successful command was:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/json-upstream bash stage3/apply.sh /workspace/scratch/json-control > /workspace/scratch/apply-retry.log 2>&1
```

Install the adapted tree's locked packages with npm ci and run npm run build. Refresh stock declarations with stock 6.0.3 `tsc -b src --force --pretty false`, then run the listed project checks. Copy the completed tree to scratch preserving symlinks. Hand-edit only the 16 annotations in the before/after table and remove the two internal markers from 43's type definitions. The proposal's normal npm build rejects the readers, as expected. One initial proposal build also hit a copied package-symlink error; after restoring all 25 package symlinks, the retry reached the actual compiler errors. That setup error is not counted as a reader error.

For the separate runtime/API experiment, stock 6.0.3 `tsc -b src/typescript --force --noCheck --pretty false` emits declarations despite the intentionally missing reader checks. Upstream's bundler command is:

```sh
node scripts/dtsBundler.mjs --entrypoint built/local/typescript/typescript.d.ts --output built/local/typescript.d.ts
```

Upstream local bundling with --no-typecheck still schedules declaration typechecking and fails, but emits its runtime bundles. `npm run build:tests:notypecheck` succeeds and regenerates the harness bundle. All ten actual .js artifacts, including the harness, were SHA-256 identical to control before the suites. The prepared oracle uses these actual artifacts without re-running the rejected build; its preparation record explicitly marks the proposal typecheck as failed. This is runtime baseline evidence for a declaration-only probe, not a passing normal lane or a checked adaptation.

The two generated public declaration files also pass an independent stock 6.0.3 strict declaration check with skipLibCheck false and no diagnostics.

No adaptation, runtime guard, lane manifest, public sanction, source pin or patch table is implemented. The full future JSON boundary adaptation, errors hidden behind existing any/assertions, external consumer source compatibility beyond the upstream API sample, native lowering/proofs, and the baseline effects of eventual runtime checks are not measured. The 39 diagnostics measure only this exact public declaration proposal. Scratch scripts, full logs, source copies and generated artifacts remain in `/workspace/scratch`; the attached reference and error tables retain the reviewable inventory and diagnostics.
