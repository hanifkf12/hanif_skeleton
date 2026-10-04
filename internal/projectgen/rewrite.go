package projectgen

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

const generatorModule = "golang.org/x/mod"

func rewriteProject(ctx context.Context, directory string, options Options, newModule string) error {
	moduleFile := filepath.Join(directory, "go.mod")
	data, err := os.ReadFile(moduleFile)
	if err != nil {
		return fmt.Errorf("read template go.mod: %w", err)
	}
	parsed, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return fmt.Errorf("parse template go.mod: %w", err)
	}
	if parsed.Module == nil || parsed.Module.Mod.Path == "" {
		return fmt.Errorf("template go.mod has no module declaration")
	}
	oldModule := parsed.Module.Mod.Path
	if err := parsed.AddModuleStmt(newModule); err != nil {
		return fmt.Errorf("rewrite module declaration: %w", err)
	}

	keepsGeneratorModule := false
	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		extension := filepath.Ext(relative)
		isGo := extension == ".go"
		isDocumentation := extension == ".md"
		isConfig := extension == ".yaml" || extension == ".yml" || extension == ".json" || extension == ".toml" || extension == ".ini" || extension == ".conf"
		// Migration and domain data are not service configuration.
		if strings.HasPrefix(relative, "database/") || strings.HasPrefix(relative, "internal/entity/") {
			isConfig = false
		}
		if !isGo && !isConfig && !isDocumentation && relative != ".env.example" && relative != "Makefile" {
			return nil
		}
		original, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read generated file %q: %w", relative, err)
		}
		var rewritten []byte
		switch {
		case isGo:
			rewriteIdentities := !strings.HasPrefix(relative, "database/") && !strings.HasPrefix(relative, "internal/entity/")
			var importsGeneratorModule bool
			rewritten, importsGeneratorModule, err = rewriteGo(relative, original, oldModule, newModule, options.Name, rewriteIdentities)
			keepsGeneratorModule = keepsGeneratorModule || importsGeneratorModule
		case relative == ".env.example":
			rewritten = []byte(rewriteAssignments(string(original), map[string]string{
				"NAME":       options.Name,
				"JWT_ISSUER": options.Name,
				"DB_NAME":    strings.ReplaceAll(options.Name, "-", "_"),
			}))
		case relative == "Makefile":
			text := stripMakeGenerator(string(original))
			rewritten = []byte(rewriteAssignments(text, map[string]string{"BINARY_NAME": options.Name}))
		case relative == "README.md":
			rewritten = []byte(rewriteREADME(string(original), oldModule, newModule, options))
		case isDocumentation:
			rewritten = []byte(replaceREADMEReferences(string(original), oldModule, newModule, options.GitURL))
		case isConfig:
			rewritten = []byte(rewriteConfigIdentities(string(original), options.Name))
		}
		if err != nil {
			return err
		}
		if !bytes.Equal(original, rewritten) {
			// WriteFile retains the existing file's mode, including executable bits.
			if err := os.WriteFile(path, rewritten, 0o644); err != nil {
				return fmt.Errorf("write generated file %q: %w", relative, err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// The generator sources are excluded. Do not carry their dependency into
	// generated projects unless another retained package actually imports it.
	if !keepsGeneratorModule {
		if err := parsed.DropRequire(generatorModule); err != nil {
			return fmt.Errorf("remove generator module: %w", err)
		}
		if err := removeGeneratorChecksums(filepath.Join(directory, "go.sum")); err != nil {
			return err
		}
	}
	parsed.Cleanup()
	updated, err := parsed.Format()
	if err != nil {
		return fmt.Errorf("format module declaration: %w", err)
	}
	if err := os.WriteFile(moduleFile, updated, 0o644); err != nil {
		return fmt.Errorf("write generated go.mod: %w", err)
	}
	return nil
}

type sourceEdit struct {
	start int
	end   int
	text  string
}

func rewriteGo(filename string, source []byte, oldModule, newModule, name string, rewriteIdentities bool) ([]byte, bool, error) {
	positions := token.NewFileSet()
	file, err := parser.ParseFile(positions, filename, source, parser.ParseComments)
	if err != nil {
		return nil, false, fmt.Errorf("parse template Go file %q: %w", filename, err)
	}
	var edits []sourceEdit
	imports := make(map[*ast.BasicLit]bool, len(file.Imports))
	importsGeneratorModule := false
	for _, declaration := range file.Imports {
		literal := declaration.Path
		imports[literal] = true
		path, err := strconv.Unquote(literal.Value)
		if err != nil {
			return nil, false, fmt.Errorf("parse import in %q: %w", filename, err)
		}
		if path == generatorModule || strings.HasPrefix(path, generatorModule+"/") {
			importsGeneratorModule = true
		}
		if path == oldModule || strings.HasPrefix(path, oldModule+"/") {
			value := newModule + strings.TrimPrefix(path, oldModule)
			edits = append(edits, sourceEdit{positions.Position(literal.Pos()).Offset, positions.Position(literal.End()).Offset, strconv.Quote(value)})
		}
	}
	if rewriteIdentities {
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING || imports[literal] {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				return true
			}
			if replacement, ok := serviceIdentity(value, name); ok {
				edits = append(edits, sourceEdit{positions.Position(literal.Pos()).Offset, positions.Position(literal.End()).Offset, strconv.Quote(replacement)})
			}
			return true
		})
	}
	if len(edits) == 0 {
		return source, importsGeneratorModule, nil
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	var rewritten bytes.Buffer
	previous := 0
	for _, edit := range edits {
		rewritten.Write(source[previous:edit.start])
		rewritten.WriteString(edit.text)
		previous = edit.end
	}
	rewritten.Write(source[previous:])
	return rewritten.Bytes(), importsGeneratorModule, nil
}

func removeGeneratorChecksums(path string) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read generated go.sum: %w", err)
	}
	var kept bytes.Buffer
	kept.Grow(len(data))
	for remaining := data; len(remaining) > 0; {
		end := bytes.IndexByte(remaining, '\n')
		if end < 0 {
			end = len(remaining) - 1
		}
		line := remaining[:end+1]
		if !bytes.HasPrefix(line, []byte(generatorModule+" ")) {
			kept.Write(line)
		}
		remaining = remaining[end+1:]
	}
	if kept.Len() == len(data) {
		return nil
	}
	if err := os.WriteFile(path, kept.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write generated go.sum: %w", err)
	}
	return nil
}

func serviceIdentity(value, name string) (string, bool) {
	// These are application identities, not arbitrary project-prefixed constants:
	// in particular, the existing PBKDF2 salt must remain unchanged.
	switch value {
	case "hanif-skeleton":
		return name, true
	case "hanif-skeleton-pubsub":
		return name + "-pubsub", true
	case "hanif-skeleton-http-client/1.0":
		return name + "-http-client/1.0", true
	default:
		return "", false
	}
}

func rewriteAssignments(text string, values map[string]string) string {
	lines := strings.SplitAfter(text, "\n")
	for i, line := range lines {
		left, _, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key := strings.TrimSpace(left)
		key = strings.TrimSpace(strings.TrimPrefix(key, "export "))
		key = strings.TrimSpace(strings.TrimRight(key, ":?+"))
		if value, ok := values[key]; ok {
			ending := ""
			if strings.HasSuffix(line, "\r\n") {
				ending = "\r\n"
			} else if strings.HasSuffix(line, "\n") {
				ending = "\n"
			}
			lines[i] = left + "=" + value + ending
		}
	}
	return strings.Join(lines, "")
}

func stripMakeGenerator(text string) string {
	offset := 0
	for _, line := range strings.SplitAfter(text, "\n") {
		if strings.TrimSpace(line) == "# Project generator" {
			return text[:offset]
		}
		offset += len(line)
	}
	return text
}

func stripREADMEGenerator(text string) string {
	var kept strings.Builder
	removing := false
	fence := ""
	for _, line := range strings.SplitAfter(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			marker := trimmed[:3]
			if fence == "" {
				fence = marker
			} else if marker == fence {
				fence = ""
			}
		}
		if fence == "" && trimmed == "## Project Generator" {
			removing = true
			continue
		}
		if fence == "" && strings.HasPrefix(trimmed, "## ") {
			removing = false
		}
		if !removing {
			kept.WriteString(line)
		}
	}
	return kept.String()
}

func rewriteREADME(text, oldModule, newModule string, options Options) string {
	text = stripREADMEGenerator(text)
	// Rewrite original text only once so a new URL containing the old module
	// cannot be rewritten recursively.
	text = replaceREADMEReferences(text, oldModule, newModule, options.GitURL)
	oldName := oldModule[strings.LastIndex(oldModule, "/")+1:]
	lines := strings.SplitAfter(text, "\n")
	titled := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		ending := ""
		if strings.HasSuffix(line, "\r\n") {
			ending = "\r\n"
		} else if strings.HasSuffix(line, "\n") {
			ending = "\n"
		}
		if !titled && strings.HasPrefix(line, "# ") {
			lines[i] = "# " + options.Name + ending
			titled = true
		} else if trimmed == "cd "+oldName || trimmed == "cd hanif_skeleton" {
			lines[i] = "cd " + options.Name + ending
		}
	}
	return strings.Join(lines, "")
}

func replaceREADMEReferences(text, oldModule, newModule, gitURL string) string {
	type replacement struct{ old, new string }
	replacements := []replacement{
		{DefaultTemplateURL, gitURL},
		{"https://" + oldModule + ".git", gitURL},
		{"http://" + oldModule + ".git", gitURL},
		{"https://" + oldModule, gitURL},
		{"http://" + oldModule, gitURL},
		{oldModule, newModule},
	}
	var rewritten strings.Builder
	for offset := 0; offset < len(text); {
		matched := false
		for _, candidate := range replacements {
			if !strings.HasPrefix(text[offset:], candidate.old) {
				continue
			}
			end := offset + len(candidate.old)
			if end != len(text) && text[end] != '/' && identityCharacter(text[end]) {
				continue
			}
			rewritten.WriteString(candidate.new)
			offset = end
			matched = true
			break
		}
		if !matched {
			rewritten.WriteByte(text[offset])
			offset++
		}
	}
	return rewritten.String()
}

func identityCharacter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.' || c == '/'
}

func rewriteConfigIdentities(text, name string) string {
	const prefix = "hanif-skeleton"
	var rewritten strings.Builder
	for offset := 0; offset < len(text); {
		index := strings.Index(text[offset:], prefix)
		if index < 0 {
			rewritten.WriteString(text[offset:])
			break
		}
		index += offset
		rewritten.WriteString(text[offset:index])
		matched := false
		if index == 0 || !identityCharacter(text[index-1]) {
			for _, identity := range []string{"hanif-skeleton-http-client/1.0", "hanif-skeleton-pubsub", prefix} {
				end := index + len(identity)
				if strings.HasPrefix(text[index:], identity) && (end == len(text) || !identityCharacter(text[end])) {
					replacement, _ := serviceIdentity(identity, name)
					rewritten.WriteString(replacement)
					offset = end
					matched = true
					break
				}
			}
		}
		if !matched {
			rewritten.WriteString(prefix)
			offset = index + len(prefix)
		}
	}
	return rewritten.String()
}
