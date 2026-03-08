package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type symbol struct {
	Name string
	Line int
}

type fileIndex struct {
	Path      string
	Types     []symbol
	Functions []symbol
	Variables []symbol
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fail(err)
	}

	goFiles, err := collectGoFiles(root)
	if err != nil {
		fail(err)
	}

	indexes := make([]fileIndex, 0, len(goFiles))
	for _, absPath := range goFiles {
		idx, err := indexFile(root, absPath)
		if err != nil {
			fail(err)
		}
		indexes = append(indexes, idx)
	}

	treeLines, err := buildRepositoryTree(root)
	if err != nil {
		fail(err)
	}

	if err := writeIndex(filepath.Join(root, "docs", "dev", "go_source_index.md"), treeLines, indexes); err != nil {
		fail(err)
	}
}

func collectGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func indexFile(root, absPath string) (fileIndex, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return fileIndex{}, fmt.Errorf("parse %s: %w", absPath, err)
	}

	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		return fileIndex{}, err
	}
	rel = filepath.ToSlash(rel)

	out := fileIndex{Path: rel}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			line := fset.Position(d.Pos()).Line
			if d.Recv != nil && len(d.Recv.List) > 0 {
				recvType := exprString(fset, d.Recv.List[0].Type)
				out.Functions = append(out.Functions, symbol{
					Name: fmt.Sprintf("%s (method on %s)", d.Name.Name, recvType),
					Line: line,
				})
			} else {
				out.Functions = append(out.Functions, symbol{Name: d.Name.Name, Line: line})
			}
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					out.Types = append(out.Types, symbol{
						Name: ts.Name.Name,
						Line: fset.Position(ts.Pos()).Line,
					})
				}
			case token.VAR:
				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					line := fset.Position(vs.Pos()).Line
					for _, n := range vs.Names {
						if n.Name == "_" {
							continue
						}
						out.Variables = append(out.Variables, symbol{
							Name: n.Name,
							Line: line,
						})
					}
				}
			}
		}
	}

	sortSymbols(out.Types)
	sortSymbols(out.Functions)
	sortSymbols(out.Variables)

	return out, nil
}

func exprString(fset *token.FileSet, e ast.Expr) string {
	var b bytes.Buffer
	if err := printer.Fprint(&b, fset, e); err != nil {
		return ""
	}
	return b.String()
}

func sortSymbols(v []symbol) {
	sort.Slice(v, func(i, j int) bool {
		if v[i].Line != v[j].Line {
			return v[i].Line < v[j].Line
		}
		return v[i].Name < v[j].Name
	})
}

func buildRepositoryTree(root string) ([]string, error) {
	lines := []string{filepath.Base(root) + "/"}
	more, err := buildTreeNode(root, "")
	if err != nil {
		return nil, err
	}
	lines = append(lines, more...)
	return lines, nil
}

func buildTreeNode(dir, prefix string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	filtered := make([]os.DirEntry, 0, len(entries))
	for _, ent := range entries {
		if ent.Name() == ".git" {
			continue
		}
		filtered = append(filtered, ent)
	}

	sort.Slice(filtered, func(i, j int) bool {
		li, lj := filtered[i], filtered[j]
		if li.IsDir() != lj.IsDir() {
			return li.IsDir()
		}
		return strings.ToLower(li.Name()) < strings.ToLower(lj.Name())
	})

	var lines []string
	for i, ent := range filtered {
		isLast := i == len(filtered)-1
		branch := "├── "
		nextPrefix := prefix + "│   "
		if isLast {
			branch = "└── "
			nextPrefix = prefix + "    "
		}

		name := ent.Name()
		fullPath := filepath.Join(dir, name)
		if ent.IsDir() {
			lines = append(lines, prefix+branch+name+"/")
			if name == "Rooms" {
				summary, err := summarizeRoomsTree(fullPath, nextPrefix)
				if err != nil {
					return nil, err
				}
				lines = append(lines, summary...)
				continue
			}
			child, err := buildTreeNode(fullPath, nextPrefix)
			if err != nil {
				return nil, err
			}
			lines = append(lines, child...)
			continue
		}
		lines = append(lines, prefix+branch+name)
	}

	return lines, nil
}

func summarizeRoomsTree(roomsPath, prefix string) ([]string, error) {
	entries, err := os.ReadDir(roomsPath)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		files = append(files, ent.Name())
	}
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i]) < strings.ToLower(files[j]) })

	switch len(files) {
	case 0:
		return []string{
			prefix + "└── (no files)",
			prefix + fmt.Sprintf("(%d files total)", len(files)),
		}, nil
	case 1:
		return []string{
			prefix + "├── " + files[0],
			prefix + fmt.Sprintf("(%d files total)", len(files)),
		}, nil
	default:
		return []string{
			prefix + "├── " + files[0],
			prefix + "├── ...",
			prefix + "└── " + files[len(files)-1],
			prefix + fmt.Sprintf("(%d files total)", len(files)),
		}, nil
	}
}

func writeIndex(path string, treeLines []string, indexes []fileIndex) error {
	var b strings.Builder
	b.WriteString("# Go Source Index\n\n")
	b.WriteString("Generated from current `.go` files using Go AST (`go/parser`, `go/ast`, `go/token`). ")
	b.WriteString("Use this as a quick technical map for chat/session continuity.\n\n")
	b.WriteString("## Repository Tree (Rooms summarized)\n\n")
	b.WriteString("```text\n")
	for _, line := range treeLines {
		b.WriteString(line + "\n")
	}
	b.WriteString("```\n\n")
	b.WriteString("## Go File Tree\n\n")
	for _, idx := range indexes {
		b.WriteString("- `" + idx.Path + "`\n")
	}

	for _, idx := range indexes {
		b.WriteString("\n## `" + idx.Path + "`\n\n")
		b.WriteString("Types:\n")
		writeSymbols(&b, idx.Types)
		b.WriteString("\nFunctions:\n")
		writeSymbols(&b, idx.Functions)
		b.WriteString("\nVariables:\n")
		writeSymbols(&b, idx.Variables)
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func writeSymbols(b *strings.Builder, symbols []symbol) {
	if len(symbols) == 0 {
		b.WriteString("- (none)\n")
		return
	}
	for _, s := range symbols {
		b.WriteString(fmt.Sprintf("- %s (line %d)\n", s.Name, s.Line))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "go source index generation failed:", err)
	os.Exit(1)
}
