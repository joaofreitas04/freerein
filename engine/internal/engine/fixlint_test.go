package engine_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// cli-envelope.md rule 2, enforced rather than restated: every error
// or warning diagnostic carries a fix. This walks the engine source
// and fails on any Fail(...) or Diag(Error|Warning, ...) call whose
// final argument is the literal "". Info diagnostics may omit the fix
// (rule 7's omission); severities held in variables are skipped
// conservatively — today every call site uses the constants.
func TestErrorAndWarningDiagnosticsCarryAFix(t *testing.T) {
	root := filepath.Join("..", "..") // engine/
	var violations []string
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch sel.Sel.Name {
			case "Fail":
				// e.Fail(code, message, fix) — always error severity
			case "Diag":
				// e.Diag(severity, code, message, fix)
				name := ""
				switch sev := call.Args[0].(type) {
				case *ast.SelectorExpr:
					name = sev.Sel.Name
				case *ast.Ident:
					name = sev.Name
				}
				if name != "Error" && name != "Warning" {
					return true
				}
			default:
				return true
			}
			last, ok := call.Args[len(call.Args)-1].(*ast.BasicLit)
			if ok && last.Kind == token.STRING && last.Value == `""` {
				violations = append(violations, fset.Position(call.Pos()).String())
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range violations {
		t.Errorf("error/warning diagnostic with empty fix at %s — rule 2: name the next command or edit that resolves it", v)
	}
}
