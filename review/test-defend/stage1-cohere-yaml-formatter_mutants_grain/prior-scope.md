CODE UNDER TEST: the slice's own construction and survivor guard, not Go cohere or YAML formatter correctness. ORACLE: executed Go cohere expected bytes for six wrong-output witnesses, plus self-written construction and planted-survivor assertions.

Reached local check/construction functions: formatterMutantsSetupInputs, formatterMutantsDeadline, formatterMutantsCommand, formatterMutantsOracleProduct, formatterMutantsReadCorpus, formatterMutantsUnion, formatterMutantsEnumeration, formatterMutantSurvived, formatterMutantsShard, formatterMutantsProduct, formatterMutantsBuildProduct.

Preparation callees read: formatCases, composeCases, goFormat adapter; formatCases also reaches lexCases and the repository corpus machinery. These are preparation for the check and not production-mutation targets. External Go formatter code was not edited.

Fixed plan before checking failures: W1 drop the whole survivor condition and return nil; S1 change required cases artifact name; S2 change copied source artifact names. P1/P2/P3 independently return empty values at each check/product entry. Every edit is a permitted weakened comparison, construction edit or empty-answer probe. No production mutant or native production probe is claimed.

All 26 requested names exist. Whole package has 72 top-level Tests; the matrix uses exactly the requested 26. Families and full members are in members.json. Outside-slice outcomes are unknown.
