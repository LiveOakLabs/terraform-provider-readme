//go:build tools

package tools

import (
	_ "github.com/boumenot/gocover-cobertura"                       // Test coverage reporting (make coverage).
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs" // Documentation generator (make docs).
	_ "github.com/princjef/gomarkdoc/cmd/gomarkdoc"                 // Markdown documentation generator (go generate ./...).
	_ "github.com/segmentio/golines"                                // Long line fixer.
	_ "golang.org/x/vuln/cmd/govulncheck"                           // Code vulnerability checks (make check-vuln).
	_ "mvdan.cc/gofumpt"                                            // Formatting and linting (make gofumpt).
)
