package tsprinter

// Task #c7j744j: stage1/cohere/tsprinter grain30 evidence.
// All 32 extant listed units and all 65 expression replacements passed twice under 30s.
// Action: keep all existing units. No new splits or budget exemptions.
// Measured on one instance: cpu.max=400000 100000 (4 CPU quota), Go 1.27.1, clang 20.1.8, Node 24.19.0.
// Initial origin/main: e77a4ae41f473c149aee910c51b73637686a804a.
// Rebased evidence onto origin/main 619e7a4cf33741cc04bc78dc4c0c8ba0e31d73fb; tsprinter, cohere, internal/buildcache, internal/native, internal/lower and internal sources unchanged.
// R2 restored cohere 7945d102a6c18dd36adf9114a758ce646e8b2359 before submodule commands.
// TypeScript source 050880ce59e30b356b686bd3144efe24f875ebc8; npm Prettier 3.9.6.
// One-time setup: 232.035s; go test -c -o /dev/null ./stage1/cohere/tsprinter: 203.993s.
// Products: each alone, go test ./stage1/cohere/tsprinter -run ^<product>$ -count=1 -v -timeout 10m.
// Units: each fresh process, go test ./stage1/cohere/tsprinter -run ^<unit>$ -count=1 -v -timeout 90s -test.skip ^TestProduct_.
// Go cache warm; product cache prepared by declared products; no whole test run.
// Checkpoint validation overlapped a few runs; isolated repeats replace those values below.
// Checkpoint comments changed directory product keys: initial TSC_007 wall 23.649s, expression_000 wall 25.787s, expression_060 wall 29.537s. All remained under budget; expression_060 isolated repeats are below.
// The old TestExpressionsAgainstGoAndPrettier setup unit (Loom 89.0s) is absent; _000 through _064 replace it.
// All values seconds. Setup/own entries are run 1 then run 2; command overhead is included only in walls.
// Union test does not log setup and own work separately.
//
// | Unit | Loom | Alone 1 | Alone 2 | Setup / own (1; 2) | Action |
// |---|---:|---:|---:|---|---|
// | TestMutants_025 | 125.2 | 6.570 | 6.131 | 4.203 / 0.357; 4.093 / 0.326 | keep |
// | TestMutants_020 | 111.5 | 6.274 | 6.465 | 4.060 / 0.456; 4.238 / 0.465 | keep |
// | TestTSCCorpusAgreement_006 | 108.6 | 5.281 | 5.807 | 3.144 / 0.390; 3.433 / 0.347 | keep |
// | TestTSCCorpusAgreement_002 | 95.7 | 5.292 | 5.318 | 3.096 / 0.336; 3.171 / 0.342 | keep |
// | TestMutants_021 | 84.5 | 10.033 | 7.934 | 4.380 / 0.352; 4.943 / 0.383 | keep |
// | TestMutants_024 | 78.7 | 7.570 | 7.968 | 4.531 / 0.387; 5.105 / 0.389 | keep |
// | TestMutants_002 | 76.5 | 7.094 | 6.728 | 4.461 / 0.239; 4.580 / 0.239 | keep |
// | TestMutants_017 | 67.0 | 6.168 | 6.120 | 4.055 / 0.282; 4.054 / 0.296 | keep |
// | TestMutants_018 | 59.0 | 6.206 | 6.514 | 4.060 / 0.358; 4.286 / 0.362 | keep |
// | TestMutants_011 | 58.5 | 6.320 | 6.298 | 4.225 / 0.255; 4.208 / 0.236 | keep |
// | TestMutants_006 | 58.1 | 6.162 | 6.292 | 4.207 / 0.231; 4.196 / 0.249 | keep |
// | TestMutants_013 | 57.6 | 6.252 | 6.333 | 4.202 / 0.224; 4.267 / 0.232 | keep |
// | TestMutants_014 | 57.5 | 6.148 | 6.201 | 4.082 / 0.288; 4.193 / 0.267 | keep |
// | TestMutants_012 | 56.9 | 6.054 | 6.094 | 4.077 / 0.210; 4.006 / 0.249 | keep |
// | TestMutants_004 | 56.7 | 6.148 | 6.429 | 4.186 / 0.160; 4.260 / 0.166 | keep |
// | TestTSCCorpusAgreement_007 | 55.8 | 23.649 | 5.239 | 21.352 / 0.351; 3.004 / 0.340 | keep |
// | TestMutants_022 | 54.4 | 6.347 | 6.454 | 4.113 / 0.287; 4.257 / 0.342 | keep |
// | TestMutants_008 | 54.4 | 6.437 | 6.403 | 4.195 / 0.218; 4.297 / 0.216 | keep |
// | TestMutants_016 | 54.2 | 6.644 | 6.465 | 4.434 / 0.287; 4.243 / 0.293 | keep |
// | TestMutants_005 | 51.1 | 6.378 | 6.210 | 4.310 / 0.249; 4.250 / 0.252 | keep |
// | TestMutants_028 | 48.4 | 6.082 | 6.440 | 4.169 / 0.195; 4.321 / 0.196 | keep |
// | TestMutants_009 | 48.3 | 6.156 | 6.244 | 4.183 / 0.188; 4.272 / 0.174 | keep |
// | TestMutants_007 | 47.7 | 6.282 | 6.236 | 4.262 / 0.234; 4.221 / 0.249 | keep |
// | TestMutants_019 | 44.9 | 6.316 | 6.446 | 4.149 / 0.298; 4.248 / 0.313 | keep |
// | TestTSCCorpusAgreement_Union | 44.8 | 3.978 | 3.722 | not separately logged | keep |
// | TestMutants_010 | 41.8 | 6.390 | 6.312 | 4.343 / 0.192; 4.353 / 0.205 | keep |
// | TestMutants_015 | 39.7 | 6.128 | 6.395 | 4.200 / 0.249; 4.327 / 0.280 | keep |
// | TestMutants_003 | 39.4 | 6.429 | 6.119 | 4.346 / 0.197; 4.174 / 0.203 | keep |
// | TestMutants_001 | 38.8 | 5.988 | 6.092 | 4.062 / 0.196; 4.056 / 0.209 | keep |
// | TestTSCCorpusAgreement_004 | 36.8 | 5.029 | 5.189 | 3.004 / 0.361; 3.040 / 0.344 | keep |
// | TestTSCCorpusAgreement_005 | 32.9 | 5.235 | 5.058 | 3.127 / 0.360; 2.962 / 0.346 | keep |
// | TestMutants_023 | 32.5 | 6.414 | 6.437 | 4.066 / 0.463; 4.232 / 0.474 | keep |
// | TestExpressionsAgainstGoAndPrettier_000 | — | 25.787 | 9.559 | 19.280 / 4.657; 3.386 / 4.401 | keep |
// | TestExpressionsAgainstGoAndPrettier_001 | — | 9.509 | 9.587 | 3.258 / 4.548; 3.261 / 4.462 | keep |
// | TestExpressionsAgainstGoAndPrettier_002 | — | 9.729 | 9.654 | 3.285 / 4.741; 3.319 / 4.500 | keep |
// | TestExpressionsAgainstGoAndPrettier_003 | — | 9.951 | 9.783 | 3.290 / 4.805; 3.289 / 4.552 | keep |
// | TestExpressionsAgainstGoAndPrettier_004 | — | 9.158 | 9.988 | 3.136 / 4.294; 3.334 / 4.804 | keep |
// | TestExpressionsAgainstGoAndPrettier_005 | — | 9.313 | 10.063 | 3.214 / 4.327; 3.335 / 4.916 | keep |
// | TestExpressionsAgainstGoAndPrettier_006 | — | 10.288 | 9.854 | 3.492 / 4.789; 3.239 / 4.696 | keep |
// | TestExpressionsAgainstGoAndPrettier_007 | — | 9.569 | 9.229 | 3.253 / 4.392; 3.165 / 4.287 | keep |
// | TestExpressionsAgainstGoAndPrettier_008 | — | 9.781 | 9.456 | 3.240 / 4.788; 3.172 / 4.535 | keep |
// | TestExpressionsAgainstGoAndPrettier_009 | — | 9.506 | 9.221 | 3.179 / 4.384; 3.263 / 4.266 | keep |
// | TestExpressionsAgainstGoAndPrettier_010 | — | 9.501 | 9.678 | 3.207 / 4.334; 3.290 / 4.357 | keep |
// | TestExpressionsAgainstGoAndPrettier_011 | — | 9.363 | 9.478 | 3.187 / 4.373; 3.238 / 4.454 | keep |
// | TestExpressionsAgainstGoAndPrettier_012 | — | 9.469 | 9.519 | 3.248 / 4.474; 3.269 / 4.442 | keep |
// | TestExpressionsAgainstGoAndPrettier_013 | — | 10.307 | 11.329 | 3.309 / 5.210; 3.592 / 5.928 | keep |
// | TestExpressionsAgainstGoAndPrettier_014 | — | 10.043 | 9.814 | 3.423 / 4.670; 3.491 / 4.486 | keep |
// | TestExpressionsAgainstGoAndPrettier_015 | — | 10.392 | 9.917 | 3.345 / 5.144; 3.448 / 4.584 | keep |
// | TestExpressionsAgainstGoAndPrettier_016 | — | 9.894 | 9.507 | 3.421 / 4.604; 3.260 / 4.413 | keep |
// | TestExpressionsAgainstGoAndPrettier_017 | — | 10.628 | 10.314 | 3.492 / 5.291; 3.490 / 4.792 | keep |
// | TestExpressionsAgainstGoAndPrettier_018 | — | 9.253 | 9.638 | 3.184 / 4.320; 3.275 / 4.509 | keep |
// | TestExpressionsAgainstGoAndPrettier_019 | — | 9.718 | 9.961 | 3.344 / 4.574; 3.358 / 4.764 | keep |
// | TestExpressionsAgainstGoAndPrettier_020 | — | 10.159 | 9.872 | 3.539 / 4.645; 3.537 / 4.486 | keep |
// | TestExpressionsAgainstGoAndPrettier_021 | — | 9.529 | 9.537 | 3.325 / 4.376; 3.379 / 4.353 | keep |
// | TestExpressionsAgainstGoAndPrettier_022 | — | 10.161 | 9.892 | 3.473 / 4.825; 3.314 / 4.673 | keep |
// | TestExpressionsAgainstGoAndPrettier_023 | — | 10.094 | 9.634 | 3.678 / 4.612; 3.331 / 4.563 | keep |
// | TestExpressionsAgainstGoAndPrettier_024 | — | 9.965 | 9.669 | 3.368 / 4.735; 3.258 / 4.573 | keep |
// | TestExpressionsAgainstGoAndPrettier_025 | — | 9.494 | 9.541 | 3.272 / 4.376; 3.262 / 4.554 | keep |
// | TestExpressionsAgainstGoAndPrettier_026 | — | 9.610 | 10.240 | 3.205 / 4.478; 3.543 / 4.818 | keep |
// | TestExpressionsAgainstGoAndPrettier_027 | — | 10.519 | 10.365 | 3.555 / 4.983; 3.474 / 5.033 | keep |
// | TestExpressionsAgainstGoAndPrettier_028 | — | 10.047 | 9.743 | 3.459 / 4.650; 3.455 / 4.472 | keep |
// | TestExpressionsAgainstGoAndPrettier_029 | — | 10.160 | 9.781 | 3.395 / 4.841; 3.317 / 4.616 | keep |
// | TestExpressionsAgainstGoAndPrettier_030 | — | 9.694 | 10.050 | 3.477 / 4.431; 3.474 / 4.598 | keep |
// | TestExpressionsAgainstGoAndPrettier_031 | — | 10.639 | 9.858 | 3.599 / 5.189; 3.388 / 4.683 | keep |
// | TestExpressionsAgainstGoAndPrettier_032 | — | 9.543 | 9.808 | 3.252 / 4.529; 3.348 / 4.621 | keep |
// | TestExpressionsAgainstGoAndPrettier_033 | — | 9.599 | 9.955 | 3.459 / 4.249; 3.374 / 4.844 | keep |
// | TestExpressionsAgainstGoAndPrettier_034 | — | 10.012 | 9.481 | 3.475 / 4.636; 3.372 / 4.340 | keep |
// | TestExpressionsAgainstGoAndPrettier_035 | — | 9.724 | 9.792 | 3.301 / 4.613; 3.438 / 4.440 | keep |
// | TestExpressionsAgainstGoAndPrettier_036 | — | 9.532 | 9.630 | 3.313 / 4.418; 3.282 / 4.435 | keep |
// | TestExpressionsAgainstGoAndPrettier_037 | — | 9.620 | 9.962 | 3.184 / 4.671; 3.324 / 4.780 | keep |
// | TestExpressionsAgainstGoAndPrettier_038 | — | 10.014 | 9.672 | 3.306 / 4.949; 3.194 / 4.559 | keep |
// | TestExpressionsAgainstGoAndPrettier_039 | — | 9.864 | 9.920 | 3.429 / 4.626; 3.300 / 4.608 | keep |
// | TestExpressionsAgainstGoAndPrettier_040 | — | 9.648 | 9.698 | 3.294 / 4.561; 3.251 / 4.527 | keep |
// | TestExpressionsAgainstGoAndPrettier_041 | — | 11.435 | 11.302 | 3.482 / 6.078; 3.394 / 5.976 | keep |
// | TestExpressionsAgainstGoAndPrettier_042 | — | 9.283 | 9.233 | 3.154 / 4.377; 3.107 / 4.397 | keep |
// | TestExpressionsAgainstGoAndPrettier_043 | — | 9.532 | 9.756 | 3.115 / 4.629; 3.314 / 4.617 | keep |
// | TestExpressionsAgainstGoAndPrettier_044 | — | 10.501 | 11.567 | 3.277 / 5.467; 3.405 / 6.176 | keep |
// | TestExpressionsAgainstGoAndPrettier_045 | — | 10.428 | 10.512 | 3.454 / 5.024; 3.582 / 5.014 | keep |
// | TestExpressionsAgainstGoAndPrettier_046 | — | 10.485 | 9.893 | 3.712 / 4.831; 3.466 / 4.528 | keep |
// | TestExpressionsAgainstGoAndPrettier_047 | — | 9.408 | 9.738 | 3.276 / 4.423; 3.316 / 4.588 | keep |
// | TestExpressionsAgainstGoAndPrettier_048 | — | 10.182 | 10.302 | 3.286 / 5.126; 3.301 / 5.336 | keep |
// | TestExpressionsAgainstGoAndPrettier_049 | — | 9.553 | 9.240 | 3.352 / 4.342; 3.086 / 4.411 | keep |
// | TestExpressionsAgainstGoAndPrettier_050 | — | 10.574 | 10.655 | 3.362 / 5.367; 3.386 / 5.579 | keep |
// | TestExpressionsAgainstGoAndPrettier_051 | — | 9.841 | 10.306 | 3.416 / 4.655; 3.431 / 4.961 | keep |
// | TestExpressionsAgainstGoAndPrettier_052 | — | 9.692 | 9.591 | 3.330 / 4.567; 3.289 / 4.517 | keep |
// | TestExpressionsAgainstGoAndPrettier_053 | — | 9.863 | 9.291 | 3.286 / 4.717; 3.150 / 4.383 | keep |
// | TestExpressionsAgainstGoAndPrettier_054 | — | 9.796 | 9.682 | 3.234 / 4.809; 3.267 / 4.540 | keep |
// | TestExpressionsAgainstGoAndPrettier_055 | — | 9.535 | 9.282 | 3.170 / 4.574; 3.149 / 4.354 | keep |
// | TestExpressionsAgainstGoAndPrettier_056 | — | 9.256 | 9.361 | 3.177 / 4.390; 3.198 / 4.411 | keep |
// | TestExpressionsAgainstGoAndPrettier_057 | — | 9.372 | 9.323 | 3.380 / 4.048; 3.302 / 4.185 | keep |
// | TestExpressionsAgainstGoAndPrettier_058 | — | 11.288 | 10.267 | 3.985 / 5.266; 3.536 / 4.822 | keep |
// | TestExpressionsAgainstGoAndPrettier_059 | — | 10.108 | 10.320 | 3.505 / 4.761; 3.576 / 4.823 | keep |
// | TestExpressionsAgainstGoAndPrettier_060 | — | 10.353 | 10.299 | 3.574 / 4.752; 3.747 / 4.593 | keep |
// | TestExpressionsAgainstGoAndPrettier_061 | — | 9.957 | 9.925 | 3.277 / 4.838; 3.327 / 4.728 | keep |
// | TestExpressionsAgainstGoAndPrettier_062 | — | 11.506 | 11.268 | 3.458 / 6.186; 3.302 / 6.104 | keep |
// | TestExpressionsAgainstGoAndPrettier_063 | — | 11.253 | 10.275 | 4.497 / 4.890; 3.512 / 4.804 | keep |
// | TestExpressionsAgainstGoAndPrettier_064 | — | 5.749 | 5.892 | 3.348 / 0.378; 3.494 / 0.352 | keep |
//
// Build phase product walls (first successful observed build; excludes initial corpus skip):
// | Product | Wall | Exit |
// |---|---:|---:|
// | TestProduct_TSPrinterExpressionsCorpus | 8.676 | 0 |
// | TestProduct_TSPrinterExpressionsLowered | 19.739 | 0 |
// | TestProduct_TSPrinterExpressionsRelease | 14.218 | 0 |
// | TestProduct_TSPrinterExpressionsSanitized | 17.717 | 0 |
// | TestProduct_TSPrinterGoOracle | 50.691 | 0 |
// | TestProduct_TSPrinterMutantNative_000 | 14.055 | 0 |
// | TestProduct_TSPrinterMutantNative_001 | 30.311 | 0 |
// | TestProduct_TSPrinterMutantNative_002 | 30.091 | 0 |
// | TestProduct_TSPrinterMutantNative_003 | 30.573 | 0 |
// | TestProduct_TSPrinterMutantNative_004 | 31.447 | 0 |
// | TestProduct_TSPrinterMutantNative_005 | 23.078 | 0 |
// | TestProduct_TSPrinterMutantNative_006 | 24.154 | 0 |
// | TestProduct_TSPrinterMutantNative_007 | 32.539 | 0 |
// | TestProduct_TSPrinterMutantNative_008 | 23.513 | 0 |
// | TestProduct_TSPrinterMutantNative_009 | 21.727 | 0 |
// | TestProduct_TSPrinterMutantNative_010 | 30.635 | 0 |
// | TestProduct_TSPrinterMutantNative_011 | 30.083 | 0 |
// | TestProduct_TSPrinterMutantNative_012 | 26.419 | 0 |
// | TestProduct_TSPrinterMutantNative_013 | 31.975 | 0 |
// | TestProduct_TSPrinterMutantNative_014 | 28.448 | 0 |
// | TestProduct_TSPrinterMutantNative_015 | 23.252 | 0 |
// | TestProduct_TSPrinterMutantNative_016 | 22.113 | 0 |
// | TestProduct_TSPrinterMutantNative_017 | 26.007 | 0 |
// | TestProduct_TSPrinterMutantNative_018 | 22.330 | 0 |
// | TestProduct_TSPrinterMutantNative_019 | 26.453 | 0 |
// | TestProduct_TSPrinterMutantNative_020 | 23.222 | 0 |
// | TestProduct_TSPrinterMutantNative_021 | 29.065 | 0 |
// | TestProduct_TSPrinterMutantNative_022 | 23.811 | 0 |
// | TestProduct_TSPrinterMutantNative_023 | 26.600 | 0 |
// | TestProduct_TSPrinterMutantNative_024 | 30.628 | 0 |
// | TestProduct_TSPrinterMutantNative_025 | 22.384 | 0 |
// | TestProduct_TSPrinterMutantNative_026 | 8.526 | 0 |
// | TestProduct_TSPrinterMutantNative_027 | 5.801 | 0 |
// | TestProduct_TSPrinterMutantNative_028 | 30.413 | 0 |
// | TestProduct_TSPrinterMutantOracle_000 | 10.146 | 0 |
// | TestProduct_TSPrinterMutantOracle_001 | 11.841 | 0 |
// | TestProduct_TSPrinterMutantOracle_002 | 4.768 | 0 |
// | TestProduct_TSPrinterMutantOracle_003 | 8.699 | 0 |
// | TestProduct_TSPrinterMutantOracle_004 | 8.721 | 0 |
// | TestProduct_TSPrinterMutantOracle_005 | 8.643 | 0 |
// | TestProduct_TSPrinterMutantOracle_006 | 8.744 | 0 |
// | TestProduct_TSPrinterMutantOracle_007 | 8.712 | 0 |
// | TestProduct_TSPrinterMutantOracle_008 | 8.671 | 0 |
// | TestProduct_TSPrinterMutantOracle_009 | 9.613 | 0 |
// | TestProduct_TSPrinterMutantOracle_010 | 10.163 | 0 |
// | TestProduct_TSPrinterMutantOracle_011 | 10.061 | 0 |
// | TestProduct_TSPrinterMutantOracle_012 | 9.808 | 0 |
// | TestProduct_TSPrinterMutantOracle_013 | 9.720 | 0 |
// | TestProduct_TSPrinterMutantOracle_014 | 9.543 | 0 |
// | TestProduct_TSPrinterMutantOracle_015 | 9.148 | 0 |
// | TestProduct_TSPrinterMutantOracle_016 | 5.124 | 0 |
// | TestProduct_TSPrinterMutantOracle_017 | 5.070 | 0 |
// | TestProduct_TSPrinterMutantOracle_018 | 9.286 | 0 |
// | TestProduct_TSPrinterMutantOracle_019 | 9.467 | 0 |
// | TestProduct_TSPrinterMutantOracle_020 | 9.337 | 0 |
// | TestProduct_TSPrinterMutantOracle_021 | 9.300 | 0 |
// | TestProduct_TSPrinterMutantOracle_022 | 9.427 | 0 |
// | TestProduct_TSPrinterMutantOracle_023 | 9.148 | 0 |
// | TestProduct_TSPrinterMutantOracle_024 | 9.181 | 0 |
// | TestProduct_TSPrinterMutantOracle_025 | 8.947 | 0 |
// | TestProduct_TSPrinterMutantOracle_026 | 4.375 | 0 |
// | TestProduct_TSPrinterMutantOracle_027 | 4.492 | 0 |
// | TestProduct_TSPrinterMutantOracle_028 | 9.052 | 0 |
// | TestProduct_TSPrinterStatementsCorpus | 5.850 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_000 | 5.431 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_001 | 5.529 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_002 | 5.479 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_003 | 5.555 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_004 | 5.559 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_005 | 5.631 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_006 | 6.053 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_007 | 6.047 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_008 | 5.330 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_009 | 5.441 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_010 | 5.487 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_011 | 5.153 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_012 | 5.217 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_013 | 5.390 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_014 | 5.440 | 0 |
// | TestProduct_TSPrinterStatementsEmbedded_015 | 5.114 | 0 |
// | TestProduct_TSPrinterStatementsLowered | 20.305 | 0 |
// | TestProduct_TSPrinterStatementsNPM_000 | 5.642 | 0 |
// | TestProduct_TSPrinterStatementsNPM_001 | 5.308 | 0 |
// | TestProduct_TSPrinterStatementsNPM_002 | 5.535 | 0 |
// | TestProduct_TSPrinterStatementsNPM_003 | 5.653 | 0 |
// | TestProduct_TSPrinterStatementsNPM_004 | 5.502 | 0 |
// | TestProduct_TSPrinterStatementsNPM_005 | 5.428 | 0 |
// | TestProduct_TSPrinterStatementsNPM_006 | 5.541 | 0 |
// | TestProduct_TSPrinterStatementsNPM_007 | 5.324 | 0 |
// | TestProduct_TSPrinterStatementsNPM_008 | 5.815 | 0 |
// | TestProduct_TSPrinterStatementsNPM_009 | 6.249 | 0 |
// | TestProduct_TSPrinterStatementsNPM_010 | 6.030 | 0 |
// | TestProduct_TSPrinterStatementsNPM_011 | 5.896 | 0 |
// | TestProduct_TSPrinterStatementsNPM_012 | 5.634 | 0 |
// | TestProduct_TSPrinterStatementsNPM_013 | 5.965 | 0 |
// | TestProduct_TSPrinterStatementsNPM_014 | 5.922 | 0 |
// | TestProduct_TSPrinterStatementsNPM_015 | 5.790 | 0 |
// | TestProduct_TSPrinterStatementsRelease | 10.587 | 0 |
// | TestProduct_TSPrinterStatementsSanitized | 12.927 | 0 |
// | TestProduct_TSPrinterTSCExpressionsCorpus | 4.629 | 0 |
// | TestProduct_TSPrinterTSCManifest | 4.970 | 0 |
// | TestProduct_TSPrinterTSCStatementsCorpus | 2.780 | 0 |
//
// Existing union checks also run alone twice:
// TestExpressionsAgainstGoAndPrettierUnion: 4.101s / 4.226s, both PASS.
// TestMutantsUnion: 1.969s / 1.767s, both PASS.
// No children were added, so no new planted-failure test is required.
// Raw command output remains in /tmp/tsprinter-grain30 on the measurement instance.
