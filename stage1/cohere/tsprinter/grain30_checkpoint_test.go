package tsprinter

// Scratch checkpoint for task #c7j744j; no test behavior changes.
// Base: origin/main e77a4ae41f473c149aee910c51b73637686a804a.
// R2 restored cohere 7945d102a6c18dd36adf9114a758ce646e8b2359.
// CPU quota: 400000/100000 (4 CPUs); Go 1.27.1, clang 20.1.8, Node 24.19.0.
// Setup wall: 232.035s; tsprinter go test -c -o /dev/null wall: 203.993s.
// External pins: TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8; Prettier 3.9.6.
// All products ran individually with -count=1 -v -timeout 10m.
// The initial ExpressionsCorpus skip is excluded.
// Product unit | command wall (seconds) | exit
// TestProduct_TSPrinterExpressionsCorpus | 8.676 | 0
// TestProduct_TSPrinterExpressionsLowered | 3.024 | 0
// TestProduct_TSPrinterExpressionsRelease | 3.205 | 0
// TestProduct_TSPrinterExpressionsSanitized | 2.994 | 0
// TestProduct_TSPrinterGoOracle | 3.036 | 0
// TestProduct_TSPrinterMutantNative_000 | 4.038 | 0
// TestProduct_TSPrinterMutantNative_001 | 30.311 | 0
// TestProduct_TSPrinterMutantNative_002 | 30.091 | 0
// TestProduct_TSPrinterMutantNative_003 | 30.573 | 0
// TestProduct_TSPrinterMutantNative_004 | 31.447 | 0
// TestProduct_TSPrinterMutantNative_005 | 23.078 | 0
// TestProduct_TSPrinterMutantNative_006 | 24.154 | 0
// TestProduct_TSPrinterMutantNative_007 | 32.539 | 0
// TestProduct_TSPrinterMutantNative_008 | 23.513 | 0
// TestProduct_TSPrinterMutantNative_009 | 21.727 | 0
// TestProduct_TSPrinterMutantNative_010 | 30.635 | 0
// TestProduct_TSPrinterMutantNative_011 | 30.083 | 0
// TestProduct_TSPrinterMutantNative_012 | 26.419 | 0
// TestProduct_TSPrinterMutantNative_013 | 31.975 | 0
// TestProduct_TSPrinterMutantNative_014 | 28.448 | 0
// TestProduct_TSPrinterMutantNative_015 | 23.252 | 0
// TestProduct_TSPrinterMutantNative_016 | 22.113 | 0
// TestProduct_TSPrinterMutantNative_017 | 26.007 | 0
// TestProduct_TSPrinterMutantNative_018 | 22.330 | 0
// TestProduct_TSPrinterMutantNative_019 | 26.453 | 0
// TestProduct_TSPrinterMutantNative_020 | 23.222 | 0
// TestProduct_TSPrinterMutantNative_021 | 29.065 | 0
// TestProduct_TSPrinterMutantNative_022 | 23.811 | 0
// TestProduct_TSPrinterMutantNative_023 | 26.600 | 0
// TestProduct_TSPrinterMutantNative_024 | 30.628 | 0
// TestProduct_TSPrinterMutantNative_025 | 22.384 | 0
// TestProduct_TSPrinterMutantNative_026 | 8.526 | 0
// TestProduct_TSPrinterMutantNative_027 | 5.801 | 0
// TestProduct_TSPrinterMutantNative_028 | 30.413 | 0
// TestProduct_TSPrinterMutantOracle_000 | 10.146 | 0
// TestProduct_TSPrinterMutantOracle_001 | 11.841 | 0
// TestProduct_TSPrinterMutantOracle_002 | 4.768 | 0
// TestProduct_TSPrinterMutantOracle_003 | 8.699 | 0
// TestProduct_TSPrinterMutantOracle_004 | 8.721 | 0
// TestProduct_TSPrinterMutantOracle_005 | 8.643 | 0
// TestProduct_TSPrinterMutantOracle_006 | 8.744 | 0
// TestProduct_TSPrinterMutantOracle_007 | 8.712 | 0
// TestProduct_TSPrinterMutantOracle_008 | 8.671 | 0
// TestProduct_TSPrinterMutantOracle_009 | 9.613 | 0
// TestProduct_TSPrinterMutantOracle_010 | 10.163 | 0
// TestProduct_TSPrinterMutantOracle_011 | 10.061 | 0
// TestProduct_TSPrinterMutantOracle_012 | 9.808 | 0
// TestProduct_TSPrinterMutantOracle_013 | 9.720 | 0
// TestProduct_TSPrinterMutantOracle_014 | 9.543 | 0
// TestProduct_TSPrinterMutantOracle_015 | 9.148 | 0
// TestProduct_TSPrinterMutantOracle_016 | 5.124 | 0
// TestProduct_TSPrinterMutantOracle_017 | 5.070 | 0
// TestProduct_TSPrinterMutantOracle_018 | 9.286 | 0
// TestProduct_TSPrinterMutantOracle_019 | 9.467 | 0
// TestProduct_TSPrinterMutantOracle_020 | 9.337 | 0
// TestProduct_TSPrinterMutantOracle_021 | 9.300 | 0
// TestProduct_TSPrinterMutantOracle_022 | 9.427 | 0
// TestProduct_TSPrinterMutantOracle_023 | 9.148 | 0
// TestProduct_TSPrinterMutantOracle_024 | 9.181 | 0
// TestProduct_TSPrinterMutantOracle_025 | 8.947 | 0
// TestProduct_TSPrinterMutantOracle_026 | 4.375 | 0
// TestProduct_TSPrinterMutantOracle_027 | 4.492 | 0
// TestProduct_TSPrinterMutantOracle_028 | 9.052 | 0
// TestProduct_TSPrinterStatementsCorpus | 5.850 | 0
// TestProduct_TSPrinterStatementsEmbedded_000 | 5.431 | 0
// TestProduct_TSPrinterStatementsEmbedded_001 | 5.529 | 0
// TestProduct_TSPrinterStatementsEmbedded_002 | 5.479 | 0
// TestProduct_TSPrinterStatementsEmbedded_003 | 5.555 | 0
// TestProduct_TSPrinterStatementsEmbedded_004 | 5.559 | 0
// TestProduct_TSPrinterStatementsEmbedded_005 | 5.631 | 0
// TestProduct_TSPrinterStatementsEmbedded_006 | 6.053 | 0
// TestProduct_TSPrinterStatementsEmbedded_007 | 6.047 | 0
// TestProduct_TSPrinterStatementsEmbedded_008 | 5.330 | 0
// TestProduct_TSPrinterStatementsEmbedded_009 | 5.441 | 0
// TestProduct_TSPrinterStatementsEmbedded_010 | 5.487 | 0
// TestProduct_TSPrinterStatementsEmbedded_011 | 5.153 | 0
// TestProduct_TSPrinterStatementsEmbedded_012 | 5.217 | 0
// TestProduct_TSPrinterStatementsEmbedded_013 | 5.390 | 0
// TestProduct_TSPrinterStatementsEmbedded_014 | 5.440 | 0
// TestProduct_TSPrinterStatementsEmbedded_015 | 5.114 | 0
// TestProduct_TSPrinterStatementsLowered | 20.305 | 0
// TestProduct_TSPrinterStatementsNPM_000 | 5.642 | 0
// TestProduct_TSPrinterStatementsNPM_001 | 5.308 | 0
// TestProduct_TSPrinterStatementsNPM_002 | 5.535 | 0
// TestProduct_TSPrinterStatementsNPM_003 | 5.653 | 0
// TestProduct_TSPrinterStatementsNPM_004 | 5.502 | 0
// TestProduct_TSPrinterStatementsNPM_005 | 5.428 | 0
// TestProduct_TSPrinterStatementsNPM_006 | 5.541 | 0
// TestProduct_TSPrinterStatementsNPM_007 | 5.324 | 0
// TestProduct_TSPrinterStatementsNPM_008 | 5.815 | 0
// TestProduct_TSPrinterStatementsNPM_009 | 6.249 | 0
// TestProduct_TSPrinterStatementsNPM_010 | 6.030 | 0
// TestProduct_TSPrinterStatementsNPM_011 | 5.896 | 0
// TestProduct_TSPrinterStatementsNPM_012 | 5.634 | 0
// TestProduct_TSPrinterStatementsNPM_013 | 5.965 | 0
// TestProduct_TSPrinterStatementsNPM_014 | 5.922 | 0
// TestProduct_TSPrinterStatementsNPM_015 | 5.790 | 0
// TestProduct_TSPrinterStatementsRelease | 10.587 | 0
// TestProduct_TSPrinterStatementsSanitized | 12.927 | 0
// TestProduct_TSPrinterTSCExpressionsCorpus | 4.629 | 0
// TestProduct_TSPrinterTSCManifest | 4.970 | 0
// TestProduct_TSPrinterTSCStatementsCorpus | 2.780 | 0
// Earlier first-build product walls (before setting external pins):
// TestProduct_TSPrinterExpressionsLowered | 19.739 | 0
// TestProduct_TSPrinterExpressionsRelease | 14.218 | 0
// TestProduct_TSPrinterExpressionsSanitized | 17.717 | 0
// TestProduct_TSPrinterGoOracle | 50.691 | 0
// TestProduct_TSPrinterMutantNative_000 | 14.055 | 0
// Units use fresh processes with -count=1 -v -timeout 90s -test.skip ^TestProduct_.
// Unit | Loom | trial | wall | setup | own | exit
// TestMutants_025 | 125.2 | 1 | 6.57 | 4.203 | 0.357 | 0
// TestMutants_025 | 125.2 | 2 | 6.131 | 4.093 | 0.326 | 0
// TestMutants_020 | 111.5 | 1 | 6.274 | 4.060 | 0.456 | 0
// TestMutants_020 | 111.5 | 2 | 6.465 | 4.238 | 0.465 | 0
// TestTSCCorpusAgreement_006 | 108.6 | 1 | 5.281 | 3.144 | 0.390 | 0
// TestTSCCorpusAgreement_006 | 108.6 | 2 | 5.807 | 3.433 | 0.347 | 0
// TestTSCCorpusAgreement_002 | 95.7 | 1 | 5.292 | 3.096 | 0.336 | 0
// TestTSCCorpusAgreement_002 | 95.7 | 2 | 5.318 | 3.171 | 0.342 | 0
// TestMutants_021 | 84.5 | 1 | 6.574 | 4.412 | 0.364 | 0
// TestMutants_021 | 84.5 | 2 | 9.988 | 4.204 | 0.303 | 0
// TestMutants_024 | 78.7 | 1 | 6.277 | 4.202 | 0.317 | 0
// TestMutants_024 | 78.7 | 2 | 6.354 | 4.140 | 0.318 | 0
// TestMutants_002 | 76.5 | 1 | 7.414 | 5.436 | 0.262 | 0
// TestMutants_002 | 76.5 | 2 | 6.574 | 4.199 | 0.245 | 0
// TestMutants_017 | 67.0 | 1 | 6.168 | 4.055 | 0.282 | 0
// TestMutants_017 | 67.0 | 2 | 6.12 | 4.054 | 0.296 | 0
// TestMutants_018 | 59.0 | 1 | 6.206 | 4.060 | 0.358 | 0
// TestMutants_018 | 59.0 | 2 | 6.514 | 4.286 | 0.362 | 0
// TestMutants_011 | 58.5 | 1 | 6.32 | 4.225 | 0.255 | 0
// TestMutants_011 | 58.5 | 2 | 6.298 | 4.208 | 0.236 | 0
// TestMutants_006 | 58.1 | 1 | 6.162 | 4.207 | 0.231 | 0
// TestMutants_006 | 58.1 | 2 | 6.292 | 4.196 | 0.249 | 0
// TestMutants_013 | 57.6 | 1 | 6.252 | 4.202 | 0.224 | 0
// TestMutants_013 | 57.6 | 2 | 6.333 | 4.267 | 0.232 | 0
// TestMutants_014 | 57.5 | 1 | 6.148 | 4.082 | 0.288 | 0
// TestMutants_014 | 57.5 | 2 | 6.201 | 4.193 | 0.267 | 0
// TestMutants_012 | 56.9 | 1 | 6.054 | 4.077 | 0.210 | 0
// TestMutants_012 | 56.9 | 2 | 6.094 | 4.006 | 0.249 | 0
// TestMutants_004 | 56.7 | 1 | 6.148 | 4.186 | 0.160 | 0
// TestMutants_004 | 56.7 | 2 | 6.429 | 4.260 | 0.166 | 0
// TestTSCCorpusAgreement_007 | 55.8 | 1 | 23.649 | 21.352 | 0.351 | 0
// TestTSCCorpusAgreement_007 | 55.8 | 2 | 5.239 | 3.004 | 0.340 | 0
// TestMutants_022 | 54.4 | 1 | 6.347 | 4.113 | 0.287 | 0
// TestMutants_022 | 54.4 | 2 | 6.454 | 4.257 | 0.342 | 0
// TestMutants_008 | 54.4 | 1 | 6.437 | 4.195 | 0.218 | 0
// TestMutants_008 | 54.4 | 2 | 6.403 | 4.297 | 0.216 | 0
// TestMutants_016 | 54.2 | 1 | 6.644 | 4.434 | 0.287 | 0
// TestMutants_016 | 54.2 | 2 | 6.465 | 4.243 | 0.293 | 0
// TestMutants_005 | 51.1 | 1 | 6.378 | 4.310 | 0.249 | 0
// TestMutants_005 | 51.1 | 2 | 6.21 | 4.250 | 0.252 | 0
// TestMutants_028 | 48.4 | 1 | 6.082 | 4.169 | 0.195 | 0
// TestMutants_028 | 48.4 | 2 | 6.44 | 4.321 | 0.196 | 0
// TestMutants_009 | 48.3 | 1 | 6.156 | 4.183 | 0.188 | 0
// TestMutants_009 | 48.3 | 2 | 6.244 | 4.272 | 0.174 | 0
// TestMutants_007 | 47.7 | 1 | 6.282 | 4.262 | 0.234 | 0
// TestMutants_007 | 47.7 | 2 | 6.236 | 4.221 | 0.249 | 0
// TestMutants_019 | 44.9 | 1 | 6.316 | 4.149 | 0.298 | 0
// TestMutants_019 | 44.9 | 2 | 6.446 | 4.248 | 0.313 | 0
// TestTSCCorpusAgreement_Union | 44.8 | 1 | 3.978 | None | None | 0
// TestTSCCorpusAgreement_Union | 44.8 | 2 | 3.722 | None | None | 0
// TestMutants_010 | 41.8 | 1 | 6.39 | 4.343 | 0.192 | 0
// TestMutants_010 | 41.8 | 2 | 6.312 | 4.353 | 0.205 | 0
// TestMutants_015 | 39.7 | 1 | 6.128 | 4.200 | 0.249 | 0
// TestMutants_015 | 39.7 | 2 | 6.395 | 4.327 | 0.280 | 0
// TestMutants_003 | 39.4 | 1 | 6.429 | 4.346 | 0.197 | 0
// TestMutants_003 | 39.4 | 2 | 6.119 | 4.174 | 0.203 | 0
// TestMutants_001 | 38.8 | 1 | 5.988 | 4.062 | 0.196 | 0
// TestMutants_001 | 38.8 | 2 | 6.092 | 4.056 | 0.209 | 0
// TestTSCCorpusAgreement_004 | 36.8 | 1 | 5.029 | 3.004 | 0.361 | 0
// TestTSCCorpusAgreement_004 | 36.8 | 2 | 5.189 | 3.040 | 0.344 | 0
// TestTSCCorpusAgreement_005 | 32.9 | 1 | 5.235 | 3.127 | 0.360 | 0
// TestTSCCorpusAgreement_005 | 32.9 | 2 | 5.058 | 2.962 | 0.346 | 0
// TestMutants_023 | 32.5 | 1 | 6.414 | 4.066 | 0.463 | 0
// TestMutants_023 | 32.5 | 2 | 6.437 | 4.232 | 0.474 | 0
// TestExpressionsAgainstGoAndPrettier_000 | None | 1 | 25.787 | 19.280 | 4.657 | 0
// TestExpressionsAgainstGoAndPrettier_000 | None | 2 | 9.559 | 3.386 | 4.401 | 0
// TestExpressionsAgainstGoAndPrettier_001 | None | 1 | 9.509 | 3.258 | 4.548 | 0
// TestExpressionsAgainstGoAndPrettier_001 | None | 2 | 9.587 | 3.261 | 4.462 | 0
// TestExpressionsAgainstGoAndPrettier_002 | None | 1 | 9.729 | 3.285 | 4.741 | 0
// TestExpressionsAgainstGoAndPrettier_002 | None | 2 | 9.654 | 3.319 | 4.500 | 0
// TestExpressionsAgainstGoAndPrettier_003 | None | 1 | 9.951 | 3.290 | 4.805 | 0
// TestExpressionsAgainstGoAndPrettier_003 | None | 2 | 9.783 | 3.289 | 4.552 | 0
// TestExpressionsAgainstGoAndPrettier_004 | None | 1 | 9.158 | 3.136 | 4.294 | 0
// TestExpressionsAgainstGoAndPrettier_004 | None | 2 | 9.988 | 3.334 | 4.804 | 0
// TestExpressionsAgainstGoAndPrettier_005 | None | 1 | 9.313 | 3.214 | 4.327 | 0
// TestExpressionsAgainstGoAndPrettier_005 | None | 2 | 10.063 | 3.335 | 4.916 | 0
// TestExpressionsAgainstGoAndPrettier_006 | None | 1 | 10.288 | 3.492 | 4.789 | 0
// TestExpressionsAgainstGoAndPrettier_006 | None | 2 | 9.854 | 3.239 | 4.696 | 0
// TestExpressionsAgainstGoAndPrettier_007 | None | 1 | 9.569 | 3.253 | 4.392 | 0
// TestExpressionsAgainstGoAndPrettier_007 | None | 2 | 9.229 | 3.165 | 4.287 | 0
// TestExpressionsAgainstGoAndPrettier_008 | None | 1 | 9.781 | 3.240 | 4.788 | 0
// TestExpressionsAgainstGoAndPrettier_008 | None | 2 | 9.456 | 3.172 | 4.535 | 0
// TestExpressionsAgainstGoAndPrettier_009 | None | 1 | 9.506 | 3.179 | 4.384 | 0
// TestExpressionsAgainstGoAndPrettier_009 | None | 2 | 9.221 | 3.263 | 4.266 | 0
// TestExpressionsAgainstGoAndPrettier_010 | None | 1 | 9.501 | 3.207 | 4.334 | 0
// TestExpressionsAgainstGoAndPrettier_010 | None | 2 | 9.678 | 3.290 | 4.357 | 0
// TestExpressionsAgainstGoAndPrettier_011 | None | 1 | 9.363 | 3.187 | 4.373 | 0
// TestExpressionsAgainstGoAndPrettier_011 | None | 2 | 9.478 | 3.238 | 4.454 | 0
// TestExpressionsAgainstGoAndPrettier_012 | None | 1 | 9.469 | 3.248 | 4.474 | 0
// TestExpressionsAgainstGoAndPrettier_012 | None | 2 | 9.519 | 3.269 | 4.442 | 0
// TestExpressionsAgainstGoAndPrettier_013 | None | 1 | 10.307 | 3.309 | 5.210 | 0
// TestExpressionsAgainstGoAndPrettier_013 | None | 2 | 11.329 | 3.592 | 5.928 | 0
// TestExpressionsAgainstGoAndPrettier_014 | None | 1 | 10.043 | 3.423 | 4.670 | 0
// TestExpressionsAgainstGoAndPrettier_014 | None | 2 | 9.814 | 3.491 | 4.486 | 0
// TestExpressionsAgainstGoAndPrettier_015 | None | 1 | 10.392 | 3.345 | 5.144 | 0
// TestExpressionsAgainstGoAndPrettier_015 | None | 2 | 9.917 | 3.448 | 4.584 | 0
// TestExpressionsAgainstGoAndPrettier_016 | None | 1 | 9.894 | 3.421 | 4.604 | 0
// TestExpressionsAgainstGoAndPrettier_016 | None | 2 | 9.507 | 3.260 | 4.413 | 0
// TestExpressionsAgainstGoAndPrettier_017 | None | 1 | 10.628 | 3.492 | 5.291 | 0
// TestExpressionsAgainstGoAndPrettier_017 | None | 2 | 10.314 | 3.490 | 4.792 | 0
// TestExpressionsAgainstGoAndPrettier_018 | None | 1 | 9.253 | 3.184 | 4.320 | 0
// TestExpressionsAgainstGoAndPrettier_018 | None | 2 | 9.638 | 3.275 | 4.509 | 0
// TestExpressionsAgainstGoAndPrettier_019 | None | 1 | 9.718 | 3.344 | 4.574 | 0
// TestExpressionsAgainstGoAndPrettier_019 | None | 2 | 9.961 | 3.358 | 4.764 | 0
// TestExpressionsAgainstGoAndPrettier_020 | None | 1 | 10.159 | 3.539 | 4.645 | 0
// TestExpressionsAgainstGoAndPrettier_020 | None | 2 | 9.872 | 3.537 | 4.486 | 0
// TestExpressionsAgainstGoAndPrettier_021 | None | 1 | 9.529 | 3.325 | 4.376 | 0
// TestExpressionsAgainstGoAndPrettier_021 | None | 2 | 9.537 | 3.379 | 4.353 | 0
// TestExpressionsAgainstGoAndPrettier_022 | None | 1 | 10.161 | 3.473 | 4.825 | 0
// TestExpressionsAgainstGoAndPrettier_022 | None | 2 | 9.892 | 3.314 | 4.673 | 0
// TestExpressionsAgainstGoAndPrettier_023 | None | 1 | 10.094 | 3.678 | 4.612 | 0
// TestExpressionsAgainstGoAndPrettier_023 | None | 2 | 9.634 | 3.331 | 4.563 | 0
// TestExpressionsAgainstGoAndPrettier_024 | None | 1 | 9.965 | 3.368 | 4.735 | 0
// TestExpressionsAgainstGoAndPrettier_024 | None | 2 | 9.669 | 3.258 | 4.573 | 0
// TestExpressionsAgainstGoAndPrettier_025 | None | 1 | 9.494 | 3.272 | 4.376 | 0
// TestExpressionsAgainstGoAndPrettier_025 | None | 2 | 9.541 | 3.262 | 4.554 | 0
// TestExpressionsAgainstGoAndPrettier_026 | None | 1 | 9.61 | 3.205 | 4.478 | 0
// TestExpressionsAgainstGoAndPrettier_026 | None | 2 | 10.24 | 3.543 | 4.818 | 0
// TestExpressionsAgainstGoAndPrettier_027 | None | 1 | 10.519 | 3.555 | 4.983 | 0
// TestExpressionsAgainstGoAndPrettier_027 | None | 2 | 10.365 | 3.474 | 5.033 | 0
// TestExpressionsAgainstGoAndPrettier_028 | None | 1 | 10.047 | 3.459 | 4.650 | 0
// TestExpressionsAgainstGoAndPrettier_028 | None | 2 | 9.743 | 3.455 | 4.472 | 0
// TestExpressionsAgainstGoAndPrettier_029 | None | 1 | 10.16 | 3.395 | 4.841 | 0
// TestExpressionsAgainstGoAndPrettier_029 | None | 2 | 9.781 | 3.317 | 4.616 | 0
// TestExpressionsAgainstGoAndPrettier_030 | None | 1 | 9.694 | 3.477 | 4.431 | 0
// TestExpressionsAgainstGoAndPrettier_030 | None | 2 | 10.05 | 3.474 | 4.598 | 0
// TestExpressionsAgainstGoAndPrettier_031 | None | 1 | 10.639 | 3.599 | 5.189 | 0
// TestExpressionsAgainstGoAndPrettier_031 | None | 2 | 9.858 | 3.388 | 4.683 | 0
// TestExpressionsAgainstGoAndPrettier_032 | None | 1 | 9.543 | 3.252 | 4.529 | 0
// TestExpressionsAgainstGoAndPrettier_032 | None | 2 | 9.808 | 3.348 | 4.621 | 0
// TestExpressionsAgainstGoAndPrettier_033 | None | 1 | 9.599 | 3.459 | 4.249 | 0
// TestExpressionsAgainstGoAndPrettier_033 | None | 2 | 9.955 | 3.374 | 4.844 | 0
// TestExpressionsAgainstGoAndPrettier_034 | None | 1 | 10.012 | 3.475 | 4.636 | 0
// TestExpressionsAgainstGoAndPrettier_034 | None | 2 | 9.481 | 3.372 | 4.340 | 0
// TestExpressionsAgainstGoAndPrettier_035 | None | 1 | 9.724 | 3.301 | 4.613 | 0
// TestExpressionsAgainstGoAndPrettier_035 | None | 2 | 9.792 | 3.438 | 4.440 | 0
// TestExpressionsAgainstGoAndPrettier_036 | None | 1 | 9.532 | 3.313 | 4.418 | 0
// TestExpressionsAgainstGoAndPrettier_036 | None | 2 | 9.63 | 3.282 | 4.435 | 0
// TestExpressionsAgainstGoAndPrettier_037 | None | 1 | 9.62 | 3.184 | 4.671 | 0
// TestExpressionsAgainstGoAndPrettier_037 | None | 2 | 9.962 | 3.324 | 4.780 | 0
// TestExpressionsAgainstGoAndPrettier_038 | None | 1 | 10.014 | 3.306 | 4.949 | 0
// TestExpressionsAgainstGoAndPrettier_038 | None | 2 | 9.672 | 3.194 | 4.559 | 0
// TestExpressionsAgainstGoAndPrettier_039 | None | 1 | 9.864 | 3.429 | 4.626 | 0
// TestExpressionsAgainstGoAndPrettier_039 | None | 2 | 9.92 | 3.300 | 4.608 | 0
// TestExpressionsAgainstGoAndPrettier_040 | None | 1 | 9.648 | 3.294 | 4.561 | 0
// TestExpressionsAgainstGoAndPrettier_040 | None | 2 | 9.698 | 3.251 | 4.527 | 0
// TestExpressionsAgainstGoAndPrettier_041 | None | 1 | 11.435 | 3.482 | 6.078 | 0
// TestExpressionsAgainstGoAndPrettier_041 | None | 2 | 11.302 | 3.394 | 5.976 | 0
// TestExpressionsAgainstGoAndPrettier_042 | None | 1 | 9.283 | 3.154 | 4.377 | 0
// TestExpressionsAgainstGoAndPrettier_042 | None | 2 | 9.233 | 3.107 | 4.397 | 0
// TestExpressionsAgainstGoAndPrettier_043 | None | 1 | 9.532 | 3.115 | 4.629 | 0
// TestExpressionsAgainstGoAndPrettier_043 | None | 2 | 9.756 | 3.314 | 4.617 | 0
// TestExpressionsAgainstGoAndPrettier_044 | None | 1 | 10.501 | 3.277 | 5.467 | 0
// TestExpressionsAgainstGoAndPrettier_044 | None | 2 | 11.567 | 3.405 | 6.176 | 0
// TestExpressionsAgainstGoAndPrettier_045 | None | 1 | 10.428 | 3.454 | 5.024 | 0
// TestExpressionsAgainstGoAndPrettier_045 | None | 2 | 10.512 | 3.582 | 5.014 | 0
// TestExpressionsAgainstGoAndPrettier_046 | None | 1 | 10.485 | 3.712 | 4.831 | 0
// TestExpressionsAgainstGoAndPrettier_046 | None | 2 | 9.893 | 3.466 | 4.528 | 0
// TestExpressionsAgainstGoAndPrettier_047 | None | 1 | 9.408 | 3.276 | 4.423 | 0
// TestExpressionsAgainstGoAndPrettier_047 | None | 2 | 9.738 | 3.316 | 4.588 | 0
// TestExpressionsAgainstGoAndPrettier_048 | None | 1 | 10.182 | 3.286 | 5.126 | 0
// TestExpressionsAgainstGoAndPrettier_048 | None | 2 | 10.302 | 3.301 | 5.336 | 0
// TestExpressionsAgainstGoAndPrettier_049 | None | 1 | 9.553 | 3.352 | 4.342 | 0
// TestExpressionsAgainstGoAndPrettier_049 | None | 2 | 9.24 | 3.086 | 4.411 | 0
// TestExpressionsAgainstGoAndPrettier_050 | None | 1 | 10.574 | 3.362 | 5.367 | 0
// TestExpressionsAgainstGoAndPrettier_050 | None | 2 | 10.655 | 3.386 | 5.579 | 0
// TestExpressionsAgainstGoAndPrettier_051 | None | 1 | 9.841 | 3.416 | 4.655 | 0
// TestExpressionsAgainstGoAndPrettier_051 | None | 2 | 10.306 | 3.431 | 4.961 | 0
// TestExpressionsAgainstGoAndPrettier_052 | None | 1 | 9.692 | 3.330 | 4.567 | 0
// TestExpressionsAgainstGoAndPrettier_052 | None | 2 | 9.591 | 3.289 | 4.517 | 0
// TestExpressionsAgainstGoAndPrettier_053 | None | 1 | 9.863 | 3.286 | 4.717 | 0
// TestExpressionsAgainstGoAndPrettier_053 | None | 2 | 9.291 | 3.150 | 4.383 | 0
// TestExpressionsAgainstGoAndPrettier_054 | None | 1 | 9.796 | 3.234 | 4.809 | 0
// TestExpressionsAgainstGoAndPrettier_054 | None | 2 | 9.682 | 3.267 | 4.540 | 0
// TestExpressionsAgainstGoAndPrettier_055 | None | 1 | 9.535 | 3.170 | 4.574 | 0
// TestExpressionsAgainstGoAndPrettier_055 | None | 2 | 9.282 | 3.149 | 4.354 | 0
// TestExpressionsAgainstGoAndPrettier_056 | None | 1 | 9.256 | 3.177 | 4.390 | 0
// TestExpressionsAgainstGoAndPrettier_056 | None | 2 | 9.361 | 3.198 | 4.411 | 0
// TestExpressionsAgainstGoAndPrettier_057 | None | 1 | 9.372 | 3.380 | 4.048 | 0
// TestExpressionsAgainstGoAndPrettier_057 | None | 2 | 9.323 | 3.302 | 4.185 | 0
// TestExpressionsAgainstGoAndPrettier_058 | None | 1 | 9.749 | 3.194 | 4.776 | 0
// TestExpressionsAgainstGoAndPrettier_058 | None | 2 | 10.699 | 3.631 | 5.028 | 0
// TestExpressionsAgainstGoAndPrettier_059 | None | 1 | 9.699 | 3.316 | 4.510 | 0
// All 32 still-existing listed units passed twice under 30s.
// The old expressions setup unit is absent; replacement measurements continue.
