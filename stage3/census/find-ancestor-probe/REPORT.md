# Probe binding census

Task 448a51k. Parent `555b1645603b90aad7b6e6df25bc642abcc52708` logs `1060360b`; head `8afae94406430e5c3340b2b33d2657615b36cf72` logs `7216e359`. Report only.

## Method

Read c6f6688c's report and both findings-only lists first. One scratch checkout of 1060360b received both complete adaptation directories from eeb87728, then `stage3/apply.sh` ran once against TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Every adapted source hash matches c6f6688c's source-hashes.json. See [carry validation](carry-validation.json), [source validation](source-validation.json), [source hashes](source-hashes.json), and [apply log](apply.log).

Both guarded entry overlays and binaries were built from their own probe pin, using the unchanged shared checker. The head change from 63c11488 to 7216e359 touches only a test fixture. See [provenance](provenance.json) and [overlays](overlays/). Entry first-error runs use `LATENT_ASSERT_NO_OUTPUT=1`, `GOMEMLIMIT=5GiB`, `GOGC=70`, `ADAMIC_PROBE_BINDINGS=1`, and separate `LATENT_TIMES` files. Neither `LATENT_FULL` nor `LATENT_ONLY_FILE` is set. Each run uses src/tsc/tsc.ts and both no-output guards. All stderr is retained whole in [parent](parent-stderr.txt.gz) and [head](head-stderr.txt.gz). Wrappers record the child's own exit, stop at 1800 seconds without output or unit progress, and clean up process groups.

The supplied lists actually contain 30 parent array refusal-scan sites and five head method-bivariance sites, not 28 plus five. All 35 listed refusal-scan sites are covered, plus three controls. Controls are the first unchanged append call in lexical order in each of es2015.ts, es2018.ts and module/module.ts, excluding changed call spans and requiring identical original findings inside the control span. Two hook assignment sites have no enclosing call expression. Their bounded assignment region is retained as an absence check, not as evidence of equal resolved signatures.

The stock TypeScript 6.0.3 AST selects the innermost call containing each finding coordinate. Selection includes every probe line whose `where` lies from that call's start through its exclusive end, including nested calls. [Site spans and source](sites.json), [extractor](sites.cjs), [analysis](analyze.py). For each site, the table below quotes each distinct full line verbatim, with occurrence counts; every occurrence remains in the filtered files in stderr order. [Parent filtered lines](parent-filtered-probes.txt), [head filtered lines](head-filtered-probes.txt), and their [parent](parent-filtered-index.json) and [head](head-filtered-index.json) stderr line indexes retain duplicates. [Independent audit](audit.json) verifies gzip integrity, unchanged source hashes, exact quotations and all 102 selected probe occurrences on each side. Same/different compares complete ordered probe-line sequences, including multiplicity. The quoted distinct lines and counts summarize that sequence without changing any line; the filtered files retain every occurrence.

## Finding stream validation

| Side | Child exit | Findings | Prior findings | Identical findings | Entire JSONL identical |
| --- | ---: | ---: | ---: | --- | --- |
| parent | 0 | 5865 | 5865 | True | True |
| head | 0 | 5834 | 5834 | True | True |

Finding comparison preserves order, multiplicity, and every field of every finding, including phase, source coordinate, reason and text. Validation compares parent against c6f6688c parent-entry-run.jsonl and head against c6f6688c head-entry-run.jsonl without path normalization. See [parent validation](parent-stream-validation.json), [head validation](head-stream-validation.json), and the raw [parent](parent-entry-run.jsonl) and [head](head-entry-run.jsonl) streams.

## Reproduction and checkpoints

Preparation and measurement commands, exact flags, own child exits, durations and silence stops are retained in the status and pid JSON files. The [setup log](setup.log) records successful setup. Controllers are [prepare.py](prepare.py) and [measure.py](measure.py). Separate checkpoint refs were pushed after preparation and each run so the final branch can have one directory-only commit directly on 1060360b without force-pushing. No main ref is pushed. Checkpoint push logs are retained losslessly as .log.gz files; their receipts retain the uncompressed hashes. [Artifact hashes](SHA256SUMS) cover the final evidence files.

## Site comparison

### 1. compiler/transformers/es2015.ts:512:13 (cleared-array)

```typescript
append(
            taggedTemplateStringDeclarations,
            factory.createVariableDeclaration(temp),
        )
```

Range: `compiler/transformers/es2015.ts:511:44` to `compiler/transformers/es2015.ts:514:10` (exclusive). Enclosing call: `compiler/transformers/es2015.ts:511:44`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2015.ts:511:44	T	resolved=(VariableDeclaration[], VariableDeclaration &#124; undefined) =&gt; VariableDeclaration[]	argument=VariableDeclaration	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2015.ts:511:44	T	resolved=(VariableDeclaration[], VariableDeclaration &#124; undefined) =&gt; VariableDeclaration[]	argument=VariableDeclaration	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 2. compiler/transformers/es2018.ts:1264:49 (cleared-array)

```typescript
append(statements, statement)
```

Range: `compiler/transformers/es2018.ts:1264:42` to `compiler/transformers/es2018.ts:1264:71` (exclusive). Enclosing call: `compiler/transformers/es2018.ts:1264:42`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:1264:42	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:1264:42	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 3. compiler/transformers/es2018.ts:1273:45 (cleared-array)

```typescript
append(statements, statement)
```

Range: `compiler/transformers/es2018.ts:1273:38` to `compiler/transformers/es2018.ts:1273:67` (exclusive). Enclosing call: `compiler/transformers/es2018.ts:1273:38`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:1273:38	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:1273:38	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 4. compiler/transformers/es2018.ts:1307:41 (cleared-array)

```typescript
append(statements, statement)
```

Range: `compiler/transformers/es2018.ts:1307:34` to `compiler/transformers/es2018.ts:1307:63` (exclusive). Enclosing call: `compiler/transformers/es2018.ts:1307:34`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:1307:34	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:1307:34	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 5. compiler/transformers/es2018.ts:1325:41 (cleared-array)

```typescript
append(statements, statement)
```

Range: `compiler/transformers/es2018.ts:1325:34` to `compiler/transformers/es2018.ts:1325:63` (exclusive). Enclosing call: `compiler/transformers/es2018.ts:1325:34`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:1325:34	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:1325:34	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 6. compiler/transformers/es2018.ts:218:13 (cleared-array)

```typescript
append(
            taggedTemplateStringDeclarations,
            factory.createVariableDeclaration(temp),
        )
```

Range: `compiler/transformers/es2018.ts:217:44` to `compiler/transformers/es2018.ts:220:10` (exclusive). Enclosing call: `compiler/transformers/es2018.ts:217:44`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:217:44	T	resolved=(VariableDeclaration[], VariableDeclaration &#124; undefined) =&gt; VariableDeclaration[]	argument=VariableDeclaration	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:217:44	T	resolved=(VariableDeclaration[], VariableDeclaration &#124; undefined) =&gt; VariableDeclaration[]	argument=VariableDeclaration	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 7. compiler/transformers/esDecorators.ts:753:41 (cleared-array)

```typescript
append(leadingBlockStatements, createMetadata(classInfo.metadataReference, classInfo.classSuper))
```

Range: `compiler/transformers/esDecorators.ts:753:34` to `compiler/transformers/esDecorators.ts:753:131` (exclusive). Enclosing call: `compiler/transformers/esDecorators.ts:753:34`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:753:34	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:753:34	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 8. compiler/transformers/esDecorators.ts:914:50 (cleared-array)

```typescript
append(trailingBlockStatements, initializerStatement)
```

Range: `compiler/transformers/esDecorators.ts:914:43` to `compiler/transformers/esDecorators.ts:914:96` (exclusive). Enclosing call: `compiler/transformers/esDecorators.ts:914:43`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:914:43	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:914:43	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 9. compiler/transformers/esDecorators.ts:924:46 (cleared-array)

```typescript
append(trailingBlockStatements, runClassInitializersStatement)
```

Range: `compiler/transformers/esDecorators.ts:924:39` to `compiler/transformers/esDecorators.ts:924:101` (exclusive). Enclosing call: `compiler/transformers/esDecorators.ts:924:39`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:924:39	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:924:39	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 10. compiler/transformers/module/esnextAnd2015.ts:266:13 (cleared-array)

```typescript
append(
            statements,
            setOriginalNode(
                setTextRange(
                    factory.createVariableStatement(
                        /*modifiers*/ undefined,
                        factory.createVariableDeclarationList(
                            [
                                factory.createVariableDeclaration(
                                    factory.cloneNode(node.name),
                                    /*exclamationToken*/ undefined,
                                    /*type*/ undefined,
                                    createRequireCall(node),
                                ),
                            ],
                            /*flags*/ languageVersion >= ScriptTarget.ES2015 ? NodeFlags.Const : NodeFlags.None,
                        ),
                    ),
                    node,
                ),
                node,
            ),
        )
```

Range: `compiler/transformers/module/esnextAnd2015.ts:265:22` to `compiler/transformers/module/esnextAnd2015.ts:287:10` (exclusive). Enclosing call: `compiler/transformers/module/esnextAnd2015.ts:265:22`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:265:22	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:267:13	T	resolved=(VariableStatement, Node &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:268:17	T	resolved=(VariableStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:274:37	T	resolved=(Identifier) =&gt; Identifier	argument=Identifier	constraint=Node &#124; undefined</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:265:22	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:267:13	T	resolved=(VariableStatement, Node &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:268:17	T	resolved=(VariableStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:274:37	T	resolved=(Identifier) =&gt; Identifier	argument=Identifier	constraint=Node &#124; undefined</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 11. compiler/transformers/module/esnextAnd2015.ts:297:17 (cleared-array)

```typescript
append(
                statements,
                factory.createExportDeclaration(
                    /*modifiers*/ undefined,
                    node.isTypeOnly,
                    factory.createNamedExports([factory.createExportSpecifier(/*isTypeOnly*/ false, /*propertyName*/ undefined, idText(node.name))]),
                ),
            )
```

Range: `compiler/transformers/module/esnextAnd2015.ts:296:26` to `compiler/transformers/module/esnextAnd2015.ts:303:14` (exclusive). Enclosing call: `compiler/transformers/module/esnextAnd2015.ts:296:26`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:296:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/esnextAnd2015.ts:296:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 12. compiler/transformers/module/module.ts:1485:21 (cleared-array)

```typescript
append(
                    statements,
                    setOriginalNode(
                        setTextRange(
                            factory.createVariableStatement(
                                /*modifiers*/ undefined,
                                factory.createVariableDeclarationList(
                                    variables,
                                    languageVersion >= ScriptTarget.ES2015 ? NodeFlags.Const : NodeFlags.None,
                                ),
                            ),
                            /*location*/ node,
                        ),
                        /*original*/ node,
                    ),
                )
```

Range: `compiler/transformers/module/module.ts:1484:30` to `compiler/transformers/module/module.ts:1499:18` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1484:30`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1484:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1486:21	T	resolved=(VariableStatement, Node &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1487:25	T	resolved=(VariableStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=TextRange</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1484:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1486:21	T	resolved=(VariableStatement, Node &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1487:25	T	resolved=(VariableStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=TextRange</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 13. compiler/transformers/module/module.ts:1505:17 (cleared-array)

```typescript
append(
                statements,
                factory.createVariableStatement(
                    /*modifiers*/ undefined,
                    factory.createVariableDeclarationList(
                        [
                            setOriginalNode(
                                setTextRange(
                                    factory.createVariableDeclaration(
                                        factory.cloneNode(namespaceDeclaration.name),
                                        /*exclamationToken*/ undefined,
                                        /*type*/ undefined,
                                        factory.getGeneratedNameForNode(node),
                                    ),
                                    /*location*/ node,
                                ),
                                /*original*/ node,
                            ),
                        ],
                        languageVersion >= ScriptTarget.ES2015 ? NodeFlags.Const : NodeFlags.None,
                    ),
                ),
            )
```

Range: `compiler/transformers/module/module.ts:1504:26` to `compiler/transformers/module/module.ts:1526:14` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1504:26`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1504:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1510:29	T	resolved=(VariableDeclaration, Node &#124; undefined) =&gt; VariableDeclaration	argument=VariableDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1511:33	T	resolved=(VariableDeclaration, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; VariableDeclaration	argument=VariableDeclaration	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1513:41	T	resolved=(Identifier) =&gt; Identifier	argument=Identifier	constraint=Node &#124; undefined</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1504:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1510:29	T	resolved=(VariableDeclaration, Node &#124; undefined) =&gt; VariableDeclaration	argument=VariableDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1511:33	T	resolved=(VariableDeclaration, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; VariableDeclaration	argument=VariableDeclaration	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1513:41	T	resolved=(Identifier) =&gt; Identifier	argument=Identifier	constraint=Node &#124; undefined</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 14. compiler/transformers/module/module.ts:1560:21 (cleared-array)

```typescript
append(
                    statements,
                    setOriginalNode(
                        setTextRange(
                            factory.createExpressionStatement(
                                createExportExpression(
                                    node.name,
                                    createRequireCall(node),
                                ),
                            ),
                            node,
                        ),
                        node,
                    ),
                )
```

Range: `compiler/transformers/module/module.ts:1559:30` to `compiler/transformers/module/module.ts:1573:18` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1559:30`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1559:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1561:21	T	resolved=(ExpressionStatement, Node &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1562:25	T	resolved=(ExpressionStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=TextRange</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1559:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1561:21	T	resolved=(ExpressionStatement, Node &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1562:25	T	resolved=(ExpressionStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=TextRange</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 15. compiler/transformers/module/module.ts:1577:21 (cleared-array)

```typescript
append(
                    statements,
                    setOriginalNode(
                        setTextRange(
                            factory.createVariableStatement(
                                /*modifiers*/ undefined,
                                factory.createVariableDeclarationList(
                                    [
                                        factory.createVariableDeclaration(
                                            factory.cloneNode(node.name),
                                            /*exclamationToken*/ undefined,
                                            /*type*/ undefined,
                                            createRequireCall(node),
                                        ),
                                    ],
                                    /*flags*/ languageVersion >= ScriptTarget.ES2015 ? NodeFlags.Const : NodeFlags.None,
                                ),
                            ),
                            node,
                        ),
                        node,
                    ),
                )
```

Range: `compiler/transformers/module/module.ts:1576:30` to `compiler/transformers/module/module.ts:1598:18` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1576:30`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1576:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1578:21	T	resolved=(VariableStatement, Node &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1579:25	T	resolved=(VariableStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1585:45	T	resolved=(Identifier) =&gt; Identifier	argument=Identifier	constraint=Node &#124; undefined</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1576:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1578:21	T	resolved=(VariableStatement, Node &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1579:25	T	resolved=(VariableStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1585:45	T	resolved=(Identifier) =&gt; Identifier	argument=Identifier	constraint=Node &#124; undefined</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 16. compiler/transformers/module/module.ts:1604:21 (cleared-array)

```typescript
append(
                    statements,
                    setOriginalNode(
                        setTextRange(
                            factory.createExpressionStatement(
                                createExportExpression(factory.getExportName(node), factory.getLocalName(node)),
                            ),
                            node,
                        ),
                        node,
                    ),
                )
```

Range: `compiler/transformers/module/module.ts:1603:30` to `compiler/transformers/module/module.ts:1614:18` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1603:30`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1603:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1605:21	T	resolved=(ExpressionStatement, Node &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1606:25	T	resolved=(ExpressionStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=TextRange</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1603:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1605:21	T	resolved=(ExpressionStatement, Node &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1606:25	T	resolved=(ExpressionStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=TextRange</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 17. compiler/transformers/module/module.ts:1753:17 (cleared-array)

```typescript
append(
                statements,
                setOriginalNode(
                    setTextRange(
                        factory.createFunctionDeclaration(
                            visitNodes(node.modifiers, modifierVisitor, isModifier),
                            node.asteriskToken,
                            factory.getDeclarationName(node, /*allowComments*/ true, /*allowSourceMaps*/ true),
                            /*typeParameters*/ undefined,
                            visitNodes(node.parameters, visitor, isParameter),
                            /*type*/ undefined,
                            visitEachChild(node.body, visitor, context),
                        ),
                        /*location*/ node,
                    ),
                    /*original*/ node,
                ),
            )
```

Range: `compiler/transformers/module/module.ts:1752:26` to `compiler/transformers/module/module.ts:1769:14` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1752:26`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1752:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1754:17	T	resolved=(FunctionDeclaration, Node &#124; undefined) =&gt; FunctionDeclaration	argument=FunctionDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1755:21	T	resolved=(FunctionDeclaration, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; FunctionDeclaration	argument=FunctionDeclaration	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1757:29	TIn	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is Modifier, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;Modifier&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1757:29	TInArray	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is Modifier, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;Modifier&gt; &#124; undefined	argument=NodeArray&lt;ModifierLike&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1757:29	TOut	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is Modifier, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;Modifier&gt; &#124; undefined	argument=Modifier	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1761:29	TIn	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1761:29	TInArray	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=NodeArray&lt;ParameterDeclaration&gt;	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1761:29	TOut	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=ParameterDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1763:29	T	resolved=(Block &#124; undefined, Visitor, TransformationContext &#124; undefined, { &lt;TIn extends Node, TInArray extends NodeArray&lt;TIn&gt; &#124; undefined, TOut extends Node&gt;(nodes: TInArray, visitor: Visitor&lt;TIn, Node &#124; undefined&gt;, test: (node: Node) =&gt; node is TOut, start?: number &#124; undefined, count?: number &#124; undefined): NodeArray&lt;...&gt; &#124; (TInArray &amp; undefined); &lt;TIn extends Node, TInArray extends Node..., Visitor &#124; undefined) =&gt; Block &#124; undefined	argument=Block	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1752:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1754:17	T	resolved=(FunctionDeclaration, Node &#124; undefined) =&gt; FunctionDeclaration	argument=FunctionDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1755:21	T	resolved=(FunctionDeclaration, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; FunctionDeclaration	argument=FunctionDeclaration	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1757:29	TIn	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is Modifier, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;Modifier&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1757:29	TInArray	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is Modifier, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;Modifier&gt; &#124; undefined	argument=NodeArray&lt;ModifierLike&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1757:29	TOut	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is Modifier, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;Modifier&gt; &#124; undefined	argument=Modifier	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1761:29	TIn	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1761:29	TInArray	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=NodeArray&lt;ParameterDeclaration&gt;	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1761:29	TOut	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=ParameterDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1763:29	T	resolved=(Block &#124; undefined, Visitor, TransformationContext &#124; undefined, { &lt;TIn extends Node, TInArray extends NodeArray&lt;TIn&gt; &#124; undefined, TOut extends Node&gt;(nodes: TInArray, visitor: Visitor&lt;TIn, Node &#124; undefined&gt;, test: (node: Node) =&gt; node is TOut, start?: number &#124; undefined, count?: number &#124; undefined): NodeArray&lt;...&gt; &#124; (TInArray &amp; undefined); &lt;TIn extends Node, TInArray extends Node..., Visitor &#124; undefined) =&gt; Block &#124; undefined	argument=Block	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 18. compiler/transformers/module/module.ts:1772:33 (cleared-array)

```typescript
append(statements, visitEachChild(node, visitor, context))
```

Range: `compiler/transformers/module/module.ts:1772:26` to `compiler/transformers/module/module.ts:1772:84` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1772:26`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1772:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1772:45	T	resolved=(FunctionDeclaration, Visitor, TransformationContext &#124; undefined) =&gt; FunctionDeclaration	argument=FunctionDeclaration	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1772:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1772:45	T	resolved=(FunctionDeclaration, Visitor, TransformationContext &#124; undefined) =&gt; FunctionDeclaration	argument=FunctionDeclaration	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 19. compiler/transformers/module/module.ts:1788:17 (cleared-array)

```typescript
append(
                statements,
                setOriginalNode(
                    setTextRange(
                        factory.createClassDeclaration(
                            visitNodes(node.modifiers, modifierVisitor, isModifierLike),
                            factory.getDeclarationName(node, /*allowComments*/ true, /*allowSourceMaps*/ true),
                            /*typeParameters*/ undefined,
                            visitNodes(node.heritageClauses, visitor, isHeritageClause),
                            visitNodes(node.members, visitor, isClassElement),
                        ),
                        node,
                    ),
                    node,
                ),
            )
```

Range: `compiler/transformers/module/module.ts:1787:26` to `compiler/transformers/module/module.ts:1802:14` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1787:26`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1787:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1789:17	T	resolved=(ClassDeclaration, Node &#124; undefined) =&gt; ClassDeclaration	argument=ClassDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1790:21	T	resolved=(ClassDeclaration, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ClassDeclaration	argument=ClassDeclaration	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1792:29	TIn	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1792:29	TInArray	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=NodeArray&lt;ModifierLike&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1792:29	TOut	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=ModifierLike	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1795:29	TIn	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1795:29	TInArray	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=NodeArray&lt;HeritageClause&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1795:29	TOut	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=HeritageClause	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1796:29	TIn	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1796:29	TInArray	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=NodeArray&lt;ClassElement&gt;	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1796:29	TOut	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=ClassElement	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1787:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1789:17	T	resolved=(ClassDeclaration, Node &#124; undefined) =&gt; ClassDeclaration	argument=ClassDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1790:21	T	resolved=(ClassDeclaration, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ClassDeclaration	argument=ClassDeclaration	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1792:29	TIn	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1792:29	TInArray	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=NodeArray&lt;ModifierLike&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1792:29	TOut	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=ModifierLike	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1795:29	TIn	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1795:29	TInArray	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=NodeArray&lt;HeritageClause&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1795:29	TOut	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=HeritageClause	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1796:29	TIn	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1796:29	TInArray	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=NodeArray&lt;ClassElement&gt;	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1796:29	TOut	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=ClassElement	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 20. compiler/transformers/module/module.ts:1805:33 (cleared-array)

```typescript
append(statements, visitEachChild(node, visitor, context))
```

Range: `compiler/transformers/module/module.ts:1805:26` to `compiler/transformers/module/module.ts:1805:84` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1805:26`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1805:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1805:45	T	resolved=(ClassDeclaration, Visitor, TransformationContext &#124; undefined) =&gt; ClassDeclaration	argument=ClassDeclaration	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1805:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1805:45	T	resolved=(ClassDeclaration, Visitor, TransformationContext &#124; undefined) =&gt; ClassDeclaration	argument=ClassDeclaration	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 21. compiler/transformers/module/module.ts:1886:37 (cleared-array)

```typescript
append(statements, factory.updateVariableStatement(node, modifiers, factory.updateVariableDeclarationList(node.declarationList, variables)))
```

Range: `compiler/transformers/module/module.ts:1886:30` to `compiler/transformers/module/module.ts:1886:170` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1886:30`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1886:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1886:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 22. compiler/transformers/module/module.ts:1894:37 (cleared-array)

```typescript
append(statements, statement)
```

Range: `compiler/transformers/module/module.ts:1894:30` to `compiler/transformers/module/module.ts:1894:59` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1894:30`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1894:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1894:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 23. compiler/transformers/module/module.ts:1898:33 (cleared-array)

```typescript
append(statements, visitEachChild(node, visitor, context))
```

Range: `compiler/transformers/module/module.ts:1898:26` to `compiler/transformers/module/module.ts:1898:84` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:1898:26`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1898:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1898:45	T	resolved=(VariableStatement, Visitor, TransformationContext &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1898:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:1898:45	T	resolved=(VariableStatement, Visitor, TransformationContext &#124; undefined) =&gt; VariableStatement	argument=VariableStatement	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 24. compiler/transformers/module/module.ts:2138:29 (cleared-array)

```typescript
append(statements, createExportStatement(exportName, expression, location, allowComments, liveBinding))
```

Range: `compiler/transformers/module/module.ts:2138:22` to `compiler/transformers/module/module.ts:2138:125` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:2138:22`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:2138:22	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:2138:22	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 25. compiler/transformers/module/system.ts:1189:29 (cleared-array)

```typescript
append(statements, createExportStatement(exportName, expression, allowComments))
```

Range: `compiler/transformers/module/system.ts:1189:22` to `compiler/transformers/module/system.ts:1189:102` (exclusive). Enclosing call: `compiler/transformers/module/system.ts:1189:22`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:1189:22	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:1189:22	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 26. compiler/transformers/module/system.ts:802:17 (cleared-array)

```typescript
append(
                hoistedStatements,
                factory.updateFunctionDeclaration(
                    node,
                    visitNodes(node.modifiers, modifierVisitor, isModifierLike),
                    node.asteriskToken,
                    factory.getDeclarationName(node, /*allowComments*/ true, /*allowSourceMaps*/ true),
                    /*typeParameters*/ undefined,
                    visitNodes(node.parameters, visitor, isParameter),
                    /*type*/ undefined,
                    visitNode(node.body, visitor, isBlock),
                ),
            )
```

Range: `compiler/transformers/module/system.ts:801:33` to `compiler/transformers/module/system.ts:813:14` (exclusive). Enclosing call: `compiler/transformers/module/system.ts:801:33`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:801:33	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:805:21	TIn	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:805:21	TInArray	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=NodeArray&lt;ModifierLike&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:805:21	TOut	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=ModifierLike	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:809:21	TIn	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:809:21	TInArray	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=NodeArray&lt;ParameterDeclaration&gt;	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:809:21	TOut	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=ParameterDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:811:21	TIn	resolved=(Block &#124; undefined, Visitor&lt;Block, Node&gt;, (node: Node) =&gt; node is Block, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; Block &#124; undefined	argument=Block &#124; undefined	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:811:21	TVisited	resolved=(Block &#124; undefined, Visitor&lt;Block, Node&gt;, (node: Node) =&gt; node is Block, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; Block &#124; undefined	argument=Node	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:811:21	TOut	resolved=(Block &#124; undefined, Visitor&lt;Block, Node&gt;, (node: Node) =&gt; node is Block, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; Block &#124; undefined	argument=Block	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:801:33	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:805:21	TIn	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:805:21	TInArray	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=NodeArray&lt;ModifierLike&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:805:21	TOut	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=ModifierLike	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:809:21	TIn	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:809:21	TInArray	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=NodeArray&lt;ParameterDeclaration&gt;	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:809:21	TOut	resolved=(NodeArray&lt;ParameterDeclaration&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ParameterDeclaration, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ParameterDeclaration&gt;	argument=ParameterDeclaration	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:811:21	TIn	resolved=(Block &#124; undefined, Visitor&lt;Block, Node&gt;, (node: Node) =&gt; node is Block, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; Block &#124; undefined	argument=Block &#124; undefined	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:811:21	TVisited	resolved=(Block &#124; undefined, Visitor&lt;Block, Node&gt;, (node: Node) =&gt; node is Block, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; Block &#124; undefined	argument=Node	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:811:21	TOut	resolved=(Block &#124; undefined, Visitor&lt;Block, Node&gt;, (node: Node) =&gt; node is Block, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; Block &#124; undefined	argument=Block	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 27. compiler/transformers/module/system.ts:816:40 (cleared-array)

```typescript
append(hoistedStatements, visitEachChild(node, visitor, context))
```

Range: `compiler/transformers/module/system.ts:816:33` to `compiler/transformers/module/system.ts:816:98` (exclusive). Enclosing call: `compiler/transformers/module/system.ts:816:33`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:816:33	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:816:59	T	resolved=(FunctionDeclaration, Visitor, TransformationContext &#124; undefined) =&gt; FunctionDeclaration	argument=FunctionDeclaration	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:816:33	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:816:59	T	resolved=(FunctionDeclaration, Visitor, TransformationContext &#124; undefined) =&gt; FunctionDeclaration	argument=FunctionDeclaration	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 28. compiler/transformers/module/system.ts:837:13 (cleared-array)

```typescript
append(
            statements,
            setTextRange(
                factory.createExpressionStatement(
                    factory.createAssignment(
                        name,
                        setTextRange(
                            factory.createClassExpression(
                                visitNodes(node.modifiers, modifierVisitor, isModifierLike),
                                node.name,
                                /*typeParameters*/ undefined,
                                visitNodes(node.heritageClauses, visitor, isHeritageClause),
                                visitNodes(node.members, visitor, isClassElement),
                            ),
                            node,
                        ),
                    ),
                ),
                node,
            ),
        )
```

Range: `compiler/transformers/module/system.ts:836:22` to `compiler/transformers/module/system.ts:856:10` (exclusive). Enclosing call: `compiler/transformers/module/system.ts:836:22`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:836:22	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:838:13	T	resolved=(ExpressionStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:842:25	T	resolved=(ClassExpression, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ClassExpression	argument=ClassExpression	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:844:33	TIn	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:844:33	TInArray	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=NodeArray&lt;ModifierLike&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:844:33	TOut	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=ModifierLike	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:847:33	TIn	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:847:33	TInArray	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=NodeArray&lt;HeritageClause&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:847:33	TOut	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=HeritageClause	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:848:33	TIn	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:848:33	TInArray	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=NodeArray&lt;ClassElement&gt;	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:848:33	TOut	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=ClassElement	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:836:22	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:838:13	T	resolved=(ExpressionStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:842:25	T	resolved=(ClassExpression, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ClassExpression	argument=ClassExpression	constraint=TextRange</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:844:33	TIn	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:844:33	TInArray	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=NodeArray&lt;ModifierLike&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:844:33	TOut	resolved=(NodeArray&lt;ModifierLike&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ModifierLike, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ModifierLike&gt; &#124; undefined	argument=ModifierLike	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:847:33	TIn	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:847:33	TInArray	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=NodeArray&lt;HeritageClause&gt; &#124; undefined	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:847:33	TOut	resolved=(NodeArray&lt;HeritageClause&gt; &#124; undefined, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is HeritageClause, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;HeritageClause&gt; &#124; undefined	argument=HeritageClause	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:848:33	TIn	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=Node	constraint=Node</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:848:33	TInArray	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=NodeArray&lt;ClassElement&gt;	constraint=NodeArray&lt;Node&gt; &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:848:33	TOut	resolved=(NodeArray&lt;ClassElement&gt;, Visitor&lt;Node, Node &#124; undefined&gt;, (node: Node) =&gt; node is ClassElement, number &#124; undefined, number &#124; undefined) =&gt; NodeArray&lt;ClassElement&gt;	argument=ClassElement	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 29. compiler/transformers/module/system.ts:894:33 (cleared-array)

```typescript
append(statements, factory.updateVariableStatement(node, modifiers, declarationList))
```

Range: `compiler/transformers/module/system.ts:894:26` to `compiler/transformers/module/system.ts:894:111` (exclusive). Enclosing call: `compiler/transformers/module/system.ts:894:26`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:894:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:894:26	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 30. compiler/transformers/module/system.ts:909:37 (cleared-array)

```typescript
append(statements, setTextRange(factory.createExpressionStatement(factory.inlineExpressions(expressions)), node))
```

Range: `compiler/transformers/module/system.ts:909:30` to `compiler/transformers/module/system.ts:909:143` (exclusive). Enclosing call: `compiler/transformers/module/system.ts:909:30`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:909:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:909:49	T	resolved=(ExpressionStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=TextRange</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:909:30	T	resolved=(Statement[] &#124; undefined, Statement) =&gt; Statement[]	argument=Statement	constraint={}</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/system.ts:909:49	T	resolved=(ExpressionStatement, Readonly&lt;TextRange&gt; &#124; undefined) =&gt; ExpressionStatement	argument=ExpressionStatement	constraint=TextRange</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 31. compiler/transformers/es2017.ts:163:26 (new-method)

```typescript
context.onEmitNode = onEmitNode
```

Range: `compiler/transformers/es2017.ts:163:5` to `compiler/transformers/es2017.ts:163:36` (exclusive). No enclosing call expression; assignment-region absence check.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td>No probe-binding lines.</td><td>No probe-binding lines.</td></tr>
</tbody>
</table>

**Verdict: same.** No binding observed on both sides; this is not a resolved-signature equality verdict.

### 32. compiler/transformers/es2018.ts:168:26 (new-method)

```typescript
context.onEmitNode = onEmitNode
```

Range: `compiler/transformers/es2018.ts:168:5` to `compiler/transformers/es2018.ts:168:36` (exclusive). No enclosing call expression; assignment-region absence check.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td>No probe-binding lines.</td><td>No probe-binding lines.</td></tr>
</tbody>
</table>

**Verdict: same.** No binding observed on both sides; this is not a resolved-signature equality verdict.

### 33. compiler/transformers/es2018.ts:499:40 (new-method)

```typescript
visitNode(e, visitor, isObjectLiteralElementLike)
```

Range: `compiler/transformers/es2018.ts:499:27` to `compiler/transformers/es2018.ts:499:76` (exclusive). Enclosing call: `compiler/transformers/es2018.ts:499:27`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:499:27	TIn	resolved=(MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration, Visitor&lt;NonNullable&lt;MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration&gt;, Node&gt;, (node: Node) =&gt; node is ObjectLiteralElementLike, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; ObjectLiteralElementLike	argument=MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:499:27	TVisited	resolved=(MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration, Visitor&lt;NonNullable&lt;MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration&gt;, Node&gt;, (node: Node) =&gt; node is ObjectLiteralElementLike, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; ObjectLiteralElementLike	argument=Node	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:499:27	TOut	resolved=(MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration, Visitor&lt;NonNullable&lt;MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration&gt;, Node&gt;, (node: Node) =&gt; node is ObjectLiteralElementLike, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; ObjectLiteralElementLike	argument=ObjectLiteralElementLike	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:499:27	TIn	resolved=(MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration, Visitor&lt;NonNullable&lt;MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration&gt;, Node&gt;, (node: Node) =&gt; node is ObjectLiteralElementLike, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; ObjectLiteralElementLike	argument=MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:499:27	TVisited	resolved=(MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration, Visitor&lt;NonNullable&lt;MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration&gt;, Node&gt;, (node: Node) =&gt; node is ObjectLiteralElementLike, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; ObjectLiteralElementLike	argument=Node	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:499:27	TOut	resolved=(MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration, Visitor&lt;NonNullable&lt;MethodDeclaration &#124; ShorthandPropertyAssignment &#124; AccessorDeclaration&gt;, Node&gt;, (node: Node) =&gt; node is ObjectLiteralElementLike, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; ObjectLiteralElementLike	argument=ObjectLiteralElementLike	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 34. compiler/transformers/esDecorators.ts:1931:42 (new-method)

```typescript
visitNode(node, visitor, isPropertyName)
```

Range: `compiler/transformers/esDecorators.ts:1931:26` to `compiler/transformers/esDecorators.ts:1931:66` (exclusive). Enclosing call: `compiler/transformers/esDecorators.ts:1931:26`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1931:26	TIn	resolved=(BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1931:26	TVisited	resolved=(BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=Node	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1931:26	TOut	resolved=(BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=PropertyName	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1931:26	TIn	resolved=(BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1931:26	TVisited	resolved=(BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=Node	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1931:26	TOut	resolved=(BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=PropertyName	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 35. compiler/transformers/esDecorators.ts:1954:32 (new-method)

```typescript
visitNode(node, visitor, isPropertyName)
```

Range: `compiler/transformers/esDecorators.ts:1954:16` to `compiler/transformers/esDecorators.ts:1954:56` (exclusive). Enclosing call: `compiler/transformers/esDecorators.ts:1954:16`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1954:16	TIn	resolved=(BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1954:16	TVisited	resolved=(BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=Node	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1954:16	TOut	resolved=(BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=PropertyName	constraint=Node</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1954:16	TIn	resolved=(BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1954:16	TVisited	resolved=(BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=Node	constraint=Node &#124; undefined</pre><br><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/esDecorators.ts:1954:16	TOut	resolved=(BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral, Visitor&lt;NonNullable&lt;BigIntLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral&gt;, Node&gt;, (node: Node) =&gt; node is PropertyName, ((node: readonly Node[]) =&gt; Node) &#124; undefined) =&gt; PropertyName	argument=PropertyName	constraint=Node</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 36. compiler/transformers/es2015.ts:2756:35 (control)

```typescript
append(assignments, assignment)
```

Range: `compiler/transformers/es2015.ts:2756:35` to `compiler/transformers/es2015.ts:2756:66` (exclusive). Enclosing call: `compiler/transformers/es2015.ts:2756:35`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2015.ts:2756:35	T	resolved=(Expression[] &#124; undefined, Expression) =&gt; Expression[]	argument=Expression	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2015.ts:2756:35	T	resolved=(Expression[] &#124; undefined, Expression) =&gt; Expression[]	argument=Expression	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 37. compiler/transformers/es2018.ts:757:17 (control)

```typescript
append(statements, node.statement)
```

Range: `compiler/transformers/es2018.ts:757:17` to `compiler/transformers/es2018.ts:757:51` (exclusive). Enclosing call: `compiler/transformers/es2018.ts:757:17`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:757:17	T	resolved=(Statement[], Statement &#124; undefined) =&gt; Statement[]	argument=Statement	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/es2018.ts:757:17	T	resolved=(Statement[], Statement &#124; undefined) =&gt; Statement[]	argument=Statement	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

### 38. compiler/transformers/module/module.ts:247:57 (control)

```typescript
append(importsAndRequiresToRewriteOrShim, node)
```

Range: `compiler/transformers/module/module.ts:247:57` to `compiler/transformers/module/module.ts:247:104` (exclusive). Enclosing call: `compiler/transformers/module/module.ts:247:57`.

<table>
<thead><tr><th>Parent</th><th>Head</th></tr></thead>
<tbody>
<tr><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:247:57	T	resolved=(CallExpression[] &#124; undefined, CallExpression) =&gt; CallExpression[]	argument=CallExpression	constraint={}</pre></td><td><strong>1 occurrence</strong><pre>probe-binding	/workspace/scratch/fa-adapted/src/compiler/transformers/module/module.ts:247:57	T	resolved=(CallExpression[] &#124; undefined, CallExpression) =&gt; CallExpression[]	argument=CallExpression	constraint={}</pre></td></tr>
</tbody>
</table>

**Verdict: same.** Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are identical.

## Interpretation

At all 33 changed call sites (30 array sites and three method-bivariance calls), parent and head have identical resolved signatures, arguments and constraints, with identical ordered probe-line sequences; the three unchanged append controls also match. The findings nevertheless differ, while each probe run reproduces its c6f6688c baseline byte for byte, so Adamic's scan reads something other than this binding. These logs show no checker resolution-order difference at the compared calls. The two remaining method-bivariance sites are hook assignments with no enclosing call and no probe lines, so this call-binding log does not decide those two sites.
