package namba

import "embed"

//go:embed testdata/evals/harness/*.json testdata/evals/harness/README.md
var embeddedEvalFixtures embed.FS
