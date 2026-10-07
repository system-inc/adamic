Merged main ce0750f28; area remains 65017b318. Change affecting execution: native startup resets inherited ignored SIGINT/SIGHUP/SIGTERM as Node does. Rule code, oracle, ledger and corpus inputs unchanged.

Fresh TestRulesAgree + TestOwnedWitnesses PASS 141.854s, all Go/Node/emitted JavaScript/native output comparisons pass. Focused internal/oracle TestASignalLeavesWhatWasPrinted PASS 25.891s, six default/inherited-ignored signal cases. Raw logs retained. Prior owned-mutant and corpus/profile evidence remains applicable; no repeated mutants, corpus, profiles or full gate claimed.
