package main

import "os"

func auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }
