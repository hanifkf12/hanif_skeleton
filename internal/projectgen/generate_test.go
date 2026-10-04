package projectgen

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
)

const fixtureModule = "github.com/hanifkf12/hanif_skeleton"

func TestGenerateMigratesModuleAndImportsWithoutChangingLookalikes(t *testing.T) {
	isolateGit(t)
	const source = `package service

import (
	root "github.com/hanifkf12/hanif_skeleton"
	child "github.com/hanifkf12/hanif_skeleton/internal/example"
	lookalike "github.com/hanifkf12/hanif_skeleton_extra/pkg"
	other "example.com/external/pkg"
)

// github.com/hanifkf12/hanif_skeleton/internal/example is documentation.
const Documentation = "github.com/hanifkf12/hanif_skeleton/internal/example"
const ServiceName = "hanif-skeleton-http-client/1.0"
const CryptoSalt = "hanif-skeleton-salt"
`
	const sum = "example.com/external v1.0.0 h1:fixture-checksum\n"
	template, commit := newTemplate(t, map[string]string{
		"service.go":    source,
		"go.mod":        "module " + fixtureModule + "\n\ngo 1.24.0\n\nrequire github.com/hanifkf12/hanif_skeleton_extra v1.0.0\n",
		"go.sum":        sum,
		"docs/usage.md": "Import `github.com/hanifkf12/hanif_skeleton/internal/example`; keep `github.com/hanifkf12/hanif_skeleton_extra/pkg`.\n",
	})
	parent := t.TempDir()
	result, err := generate(context.Background(), Options{
		Name: "new-service", GitURL: "https://git.example/team/new-service.git", OutputDir: parent, NoGit: true,
	}, template)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(parent, "new-service"), result.Directory)
	assert.Equal(t, "git.example/team/new-service", result.ModulePath)
	assert.Equal(t, commit, result.TemplateCommit)

	moduleFile, err := modfile.Parse("go.mod", []byte(readGeneratedFile(t, result.Directory, "go.mod")), nil)
	require.NoError(t, err)
	require.NotNil(t, moduleFile.Module)
	assert.Equal(t, result.ModulePath, moduleFile.Module.Mod.Path)
	require.Len(t, moduleFile.Require, 1)
	assert.Equal(t, "github.com/hanifkf12/hanif_skeleton_extra", moduleFile.Require[0].Mod.Path)
	assert.Equal(t, sum, readGeneratedFile(t, result.Directory, "go.sum"))
	assert.Equal(t, "Import `git.example/team/new-service/internal/example`; keep `github.com/hanifkf12/hanif_skeleton_extra/pkg`.\n",
		readGeneratedFile(t, result.Directory, "docs/usage.md"))

	generatedSource := readGeneratedFile(t, result.Directory, "service.go")
	parsed, err := parser.ParseFile(token.NewFileSet(), "service.go", generatedSource, parser.ParseComments)
	require.NoError(t, err)
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		require.NoError(t, err)
		imports = append(imports, path)
	}
	assert.ElementsMatch(t, []string{
		"git.example/team/new-service",
		"git.example/team/new-service/internal/example",
		"github.com/hanifkf12/hanif_skeleton_extra/pkg",
		"example.com/external/pkg",
	}, imports)
	assert.Contains(t, generatedSource, `"github.com/hanifkf12/hanif_skeleton/internal/example"`)
	assert.Contains(t, generatedSource, "// github.com/hanifkf12/hanif_skeleton/internal/example is documentation.")
	assert.Contains(t, generatedSource, `"new-service-http-client/1.0"`)
	assert.Contains(t, generatedSource, `"hanif-skeleton-salt"`)
}

func TestGenerateRemovesGeneratorOnlyDependencyAndPassesTidyCheck(t *testing.T) {
	isolateGit(t)
	template, _ := newTemplate(t, map[string]string{
		"go.mod": "module " + fixtureModule + "\n\ngo 1.24.0\n\n" + `require (
	git.example/first v0.0.0
	golang.org/x/mod v0.28.0
	git.example/last v0.0.0
)

replace (
	git.example/first => ./deps/first
	git.example/last => ./deps/last
)
`,
		"go.sum": "golang.org/x/mod v0.28.0 h1:gQBtGhjxykdjY9YhZpSlZIsbnaE2+PgjfLWUQTnoZ1U=\n" +
			"golang.org/x/mod v0.28.0/go.mod h1:yfB/L0NOf/kmEbXjzCPOx1iK1fRutOydrCMsqRhEBxI=\n",
		"internal/projectgen/tools.go": "package projectgen\n\nimport _ \"golang.org/x/mod/module\"\n",
		"main.go":                      "package main\n\nimport (\n\t\"git.example/first\"\n\t\"git.example/last\"\n)\n\nfunc main() { _ = first.Value + last.Value }\n",
		"deps/first/go.mod":            "module git.example/first\n\ngo 1.24.0\n",
		"deps/first/value.go":          "package first\n\nconst Value = 1\n",
		"deps/last/go.mod":             "module git.example/last\n\ngo 1.24.0\n",
		"deps/last/value.go":           "package last\n\nconst Value = 2\n",
	})
	result, err := generate(context.Background(), Options{
		Name: "tidy-project", GitURL: "https://git.example/team/tidy-project.git", OutputDir: t.TempDir(), NoGit: true,
	}, template)
	require.NoError(t, err)
	assertPathAbsent(t, filepath.Join(result.Directory, "internal/projectgen"))
	moduleFile, err := modfile.Parse("go.mod", []byte(readGeneratedFile(t, result.Directory, "go.mod")), nil)
	require.NoError(t, err)
	require.Len(t, moduleFile.Require, 2)
	assert.Equal(t, "git.example/first", moduleFile.Require[0].Mod.Path)
	assert.Equal(t, "git.example/last", moduleFile.Require[1].Mod.Path)
	assert.Empty(t, readGeneratedFile(t, result.Directory, "go.sum"))

	cmd := exec.Command("go", "mod", "tidy", "-diff")
	cmd.Dir = result.Directory
	cmd.Env = append(os.Environ(), "GOPROXY=off", "GOSUMDB=off", "GOWORK=off", "GOENV=off", "GOTOOLCHAIN=local")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated project must pass the inherited CI tidy check without network access: %s", output)
}

func TestGenerateRetainsModuleToolsUsedByApplication(t *testing.T) {
	isolateGit(t)
	const sum = "golang.org/x/mod v0.28.0 h1:gQBtGhjxykdjY9YhZpSlZIsbnaE2+PgjfLWUQTnoZ1U=\n" +
		"golang.org/x/mod v0.28.0/go.mod h1:yfB/L0NOf/kmEbXjzCPOx1iK1fRutOydrCMsqRhEBxI=\n"
	template, _ := newTemplate(t, map[string]string{
		"go.mod":  "module " + fixtureModule + "\n\ngo 1.24.0\n\nrequire golang.org/x/mod v0.28.0\n",
		"go.sum":  sum,
		"main.go": "package main\n\nimport \"golang.org/x/mod/module\"\n\nfunc main() { _ = module.CheckPath(\"git.example/team/app\") }\n",
	})
	result, err := generate(context.Background(), Options{
		Name: "module-tools-app", GitURL: "https://git.example/team/module-tools-app.git", OutputDir: t.TempDir(), NoGit: true,
	}, template)
	require.NoError(t, err)
	moduleFile, err := modfile.Parse("go.mod", []byte(readGeneratedFile(t, result.Directory, "go.mod")), nil)
	require.NoError(t, err)
	require.Len(t, moduleFile.Require, 1)
	assert.Equal(t, "golang.org/x/mod", moduleFile.Require[0].Mod.Path)
	assert.Equal(t, "v0.28.0", moduleFile.Require[0].Mod.Version)
	assert.Equal(t, sum, readGeneratedFile(t, result.Directory, "go.sum"))
}

func TestGenerateSupportsHostedGitURLForms(t *testing.T) {
	isolateGit(t)
	template, _ := newTemplate(t, nil)
	cases := []struct {
		name   string
		url    string
		module string
	}{
		{"https", "https://github.com/team/service.git", "github.com/team/service"},
		{"http", "http://git.example/team/service", "git.example/team/service"},
		{"ssh", "ssh://git@git.example/team/nested/service.git", "git.example/team/nested/service"},
		{"scp", "git@git.example:team/nested/service.git", "git.example/team/nested/service"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := generate(context.Background(), Options{
				Name: "1-service", GitURL: tc.url, OutputDir: t.TempDir(), NoGit: true,
			}, template)
			require.NoError(t, err)
			assert.Equal(t, tc.module, result.ModulePath)
			moduleFile, err := modfile.Parse("go.mod", []byte(readGeneratedFile(t, result.Directory, "go.mod")), nil)
			require.NoError(t, err)
			require.NotNil(t, moduleFile.Module)
			assert.Equal(t, tc.module, moduleFile.Module.Mod.Path)
		})
	}
}

func TestGenerateCreatesFreshMainRepositoryOrOmitsGit(t *testing.T) {
	isolateGit(t)
	template, firstCommit := newTemplate(t, nil)
	writeFixtureFile(t, template, "history.txt", "second template commit\n")
	fixtureGit(t, template, "add", "--all")
	fixtureGit(t, template, "commit", "-m", "second template commit")
	lastCommit := fixtureGit(t, template, "rev-parse", "HEAD")
	require.NotEqual(t, firstCommit, lastCommit)
	fixtureGit(t, template, "tag", "template-history")
	const targetURL = "git@git.example:team/fresh.git"

	for _, noGit := range []bool{false, true} {
		name := "fresh-repository"
		if noGit {
			name = "without-git"
		}
		t.Run(name, func(t *testing.T) {
			result, err := generate(context.Background(), Options{
				Name: "fresh", GitURL: targetURL, OutputDir: t.TempDir(), NoGit: noGit,
			}, template)
			require.NoError(t, err)
			assert.Equal(t, lastCommit, result.TemplateCommit)
			assert.Equal(t, "second template commit\n", readGeneratedFile(t, result.Directory, "history.txt"))
			if noGit {
				assertPathAbsent(t, filepath.Join(result.Directory, ".git"))
				return
			}

			assert.Equal(t, "refs/heads/main", fixtureGit(t, result.Directory, "symbolic-ref", "HEAD"))
			assert.Equal(t, targetURL, fixtureGit(t, result.Directory, "remote", "get-url", "origin"))
			assert.Equal(t, "origin", fixtureGit(t, result.Directory, "remote"))
			assert.Empty(t, fixtureGit(t, result.Directory, "rev-list", "--all"), "template history must not be copied or a commit created")
			assert.Empty(t, fixtureGit(t, result.Directory, "tag", "--list"))
			cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
			cmd.Dir = result.Directory
			_, err = cmd.CombinedOutput()
			assert.Error(t, err, "the generated repository must remain uncommitted")
		})
	}
}

func TestGenerateFiltersTrackedArtifactsAndCustomizesProjectIdentity(t *testing.T) {
	isolateGit(t)
	filtered := []string{
		".env", ".env.local", ".env.production", "nested/.env.test",
		".idea/workspace.xml", ".vscode/settings.json", ".zed/tasks.json",
		"bin/server", "coverage.out", "coverage.html", "nested/coverage.json",
		"server.out", "nested/server.test", "cmd/skeleton/main.go", "internal/projectgen/generate.go",
	}
	const migration = "-- hanif-skeleton is historical domain data\nINSERT INTO names VALUES ('hanif-skeleton');\n"
	const domain = "{\"name\":\"hanif-skeleton\",\"module\":\"github.com/hanifkf12/hanif_skeleton\"}\n"
	files := map[string]string{
		".env.example":                         "NAME=test\nJWT_ISSUER=hanif-skeleton\nDB_NAME=test_skeleton\nDB_PASSWORD=change-me\nPORT=9000\n",
		"Makefile":                             "BINARY_NAME=main\n\nbuild:\n\tgo build -o $(BINARY_NAME) .\n\n# Project generator\nskeleton:\n\tgo build -o bin/skeleton ./cmd/skeleton\n",
		"README.md":                            "# Hanif Skeleton\n\nClone:\n```sh\ngit clone https://github.com/hanifkf12/hanif_skeleton.git\ncd hanif_skeleton\n```\n\nModule: `github.com/hanifkf12/hanif_skeleton/internal/example`.\n\n## Project Generator\nGenerator-only help.\n### Options\nGenerator-only options.\n\n## Development\nKeep this development guide.\n",
		"cmd/http/http.go":                     "package http\n\nconst Service = \"hanif-skeleton\"\n",
		"otel-collector-config.yaml":           "processors:\n  resource:\n    attributes:\n      - key: service.name\n        value: \"hanif-skeleton\"\n        action: upsert\n",
		"database/migration/001.sql":           migration,
		"internal/entity/testdata/domain.json": domain,
	}
	for _, path := range filtered {
		files[path] = "tracked artifact that must not leak\n"
	}
	template, _ := newTemplate(t, files)
	result, err := generate(context.Background(), Options{
		Name: "my-service", GitURL: "https://git.example/team/my-service.git", OutputDir: t.TempDir(), NoGit: true,
	}, template)
	require.NoError(t, err)
	for _, path := range filtered {
		assertPathAbsent(t, filepath.Join(result.Directory, filepath.FromSlash(path)))
	}
	for _, path := range []string{".git", ".idea", ".vscode", ".zed", "bin", "cmd/skeleton", "internal/projectgen"} {
		assertPathAbsent(t, filepath.Join(result.Directory, filepath.FromSlash(path)))
	}

	environment := parseEnvironment(readGeneratedFile(t, result.Directory, ".env.example"))
	assert.Equal(t, "my-service", environment["NAME"])
	assert.Equal(t, "my-service", environment["JWT_ISSUER"])
	assert.Equal(t, "my_service", environment["DB_NAME"])
	assert.Equal(t, "change-me", environment["DB_PASSWORD"])
	assert.Equal(t, "9000", environment["PORT"])
	makefile := readGeneratedFile(t, result.Directory, "Makefile")
	assert.Contains(t, makefile, "BINARY_NAME=my-service")
	assert.Contains(t, makefile, "go build -o $(BINARY_NAME) .")
	assert.NotContains(t, makefile, "# Project generator")
	assert.NotContains(t, makefile, "./cmd/skeleton")
	readme := readGeneratedFile(t, result.Directory, "README.md")
	assert.True(t, strings.HasPrefix(readme, "# my-service\n"), "README title must use the project name")
	assert.Contains(t, readme, "git clone https://git.example/team/my-service.git")
	assert.Contains(t, readme, "cd my-service")
	assert.Contains(t, readme, "git.example/team/my-service/internal/example")
	assert.NotContains(t, readme, "Generator-only")
	assert.NotContains(t, readme, "## Project Generator")
	assert.Contains(t, readme, "## Development\nKeep this development guide.")
	assert.Contains(t, readGeneratedFile(t, result.Directory, "cmd/http/http.go"), `"my-service"`)
	assert.Contains(t, readGeneratedFile(t, result.Directory, "otel-collector-config.yaml"), `"my-service"`)
	assert.Equal(t, migration, readGeneratedFile(t, result.Directory, "database/migration/001.sql"))
	assert.Equal(t, domain, readGeneratedFile(t, result.Directory, "internal/entity/testdata/domain.json"))
}

func TestGenerateNeverOverwritesExistingTargets(t *testing.T) {
	isolateGit(t)
	template, _ := newTemplate(t, nil)
	for _, kind := range []string{"directory", "empty-directory", "file", "symlink", "dangling-symlink"} {
		t.Run(kind, func(t *testing.T) {
			parent := t.TempDir()
			target := filepath.Join(parent, "existing")
			var linkedDirectory string
			switch kind {
			case "directory":
				require.NoError(t, os.Mkdir(target, 0o755))
				writeFixtureFile(t, target, "sentinel.txt", "keep existing content\n")
			case "empty-directory":
				require.NoError(t, os.Mkdir(target, 0o755))
			case "file":
				require.NoError(t, os.WriteFile(target, []byte("keep existing file\n"), 0o600))
			case "symlink":
				linkedDirectory = t.TempDir()
				writeFixtureFile(t, linkedDirectory, "sentinel.txt", "keep linked content\n")
				require.NoError(t, os.Symlink(linkedDirectory, target))
			case "dangling-symlink":
				linkedDirectory = filepath.Join(parent, "missing-link-target")
				require.NoError(t, os.Symlink(linkedDirectory, target))
			}

			_, err := generate(context.Background(), Options{
				Name: "existing", GitURL: "https://git.example/team/existing.git", OutputDir: parent,
			}, template)
			require.Error(t, err)
			info, err := os.Lstat(target)
			require.NoError(t, err)
			switch kind {
			case "directory":
				assert.True(t, info.IsDir())
				assert.Equal(t, "keep existing content\n", readGeneratedFile(t, target, "sentinel.txt"))
				entries, err := os.ReadDir(target)
				require.NoError(t, err)
				assert.Len(t, entries, 1)
			case "empty-directory":
				assert.True(t, info.IsDir())
				entries, err := os.ReadDir(target)
				require.NoError(t, err)
				assert.Empty(t, entries)
			case "file":
				assert.True(t, info.Mode().IsRegular())
				assert.Equal(t, "keep existing file\n", readGeneratedFile(t, parent, "existing"))
			case "symlink", "dangling-symlink":
				assert.NotZero(t, info.Mode()&os.ModeSymlink)
				link, err := os.Readlink(target)
				require.NoError(t, err)
				assert.Equal(t, linkedDirectory, link)
				if kind == "symlink" {
					assert.Equal(t, "keep linked content\n", readGeneratedFile(t, linkedDirectory, "sentinel.txt"))
					entries, err := os.ReadDir(linkedDirectory)
					require.NoError(t, err)
					assert.Len(t, entries, 1)
				} else {
					assertPathAbsent(t, linkedDirectory)
				}
			}
		})
	}
}

func TestGenerateSelectsBranchTagAndCommit(t *testing.T) {
	isolateGit(t)
	template, initialCommit := newTemplate(t, map[string]string{"version.txt": "initial\n"})
	fixtureGit(t, template, "tag", "-a", "v1.0.0", "-m", "template release")
	fixtureGit(t, template, "checkout", "-b", "release")
	writeFixtureFile(t, template, "version.txt", "release\n")
	fixtureGit(t, template, "add", "--all")
	fixtureGit(t, template, "commit", "-m", "release template")
	releaseCommit := fixtureGit(t, template, "rev-parse", "HEAD")
	fixtureGit(t, template, "checkout", "main")
	writeFixtureFile(t, template, "version.txt", "main\n")
	fixtureGit(t, template, "add", "--all")
	fixtureGit(t, template, "commit", "-m", "main template")
	mainCommit := fixtureGit(t, template, "rev-parse", "HEAD")

	cases := []struct {
		name    string
		ref     string
		commit  string
		version string
	}{
		{"default-main", "", mainCommit, "main\n"},
		{"branch", "release", releaseCommit, "release\n"},
		{"annotated-tag", "v1.0.0", initialCommit, "initial\n"},
		{"commit", releaseCommit, releaseCommit, "release\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := generate(context.Background(), Options{
				Name: "selected", GitURL: "https://git.example/team/selected.git", OutputDir: t.TempDir(), TemplateRef: tc.ref, NoGit: true,
			}, template)
			require.NoError(t, err)
			assert.Equal(t, tc.commit, result.TemplateCommit)
			assert.Equal(t, tc.version, readGeneratedFile(t, result.Directory, "version.txt"))
		})
	}
}

func TestGenerateFailedFetchRemovesOnlyOwnedOutput(t *testing.T) {
	isolateGit(t)
	template, _ := newTemplate(t, nil)
	parent := t.TempDir()
	writeFixtureFile(t, parent, "sentinel.txt", "keep sibling\n")
	_, err := generate(context.Background(), Options{
		Name: "failed", GitURL: "https://git.example/team/failed.git", OutputDir: parent, TemplateRef: "refs/heads/does-not-exist",
	}, template)
	require.Error(t, err)
	assertPathAbsent(t, filepath.Join(parent, "failed"))
	assert.Equal(t, "keep sibling\n", readGeneratedFile(t, parent, "sentinel.txt"))
	entries, err := os.ReadDir(parent)
	require.NoError(t, err)
	require.Len(t, entries, 1, "failed generation must not leave staging directories")
	assert.Equal(t, "sentinel.txt", entries[0].Name())
}

func TestGenerateRejectsInvalidInputsBeforeCreatingOutput(t *testing.T) {
	isolateGit(t)
	template, _ := newTemplate(t, nil)
	cases := []struct {
		label string
		name  string
		url   string
		ref   string
	}{
		{"empty-name", "", "https://git.example/team/service.git", ""},
		{"parent-traversal", "../escaped", "https://git.example/team/service.git", ""},
		{"nested-path", "a/b", "https://git.example/team/service.git", ""},
		{"backslash-path", `a\b`, "https://git.example/team/service.git", ""},
		{"dot", ".", "https://git.example/team/service.git", ""},
		{"dot-dot", "..", "https://git.example/team/service.git", ""},
		{"option-name", "-project", "https://git.example/team/service.git", ""},
		{"leading-underscore", "_project", "https://git.example/team/service.git", ""},
		{"space-name", "two words", "https://git.example/team/service.git", ""},
		{"unicode-name", "prøject", "https://git.example/team/service.git", ""},
		{"empty-url", "service", "", ""},
		{"option-url", "service", "--upload-pack=bad", ""},
		{"http-password", "service", "https://user:password@git.example/team/service.git", ""},
		{"http-token", "service", "https://token@git.example/team/service.git", ""},
		{"url-query", "service", "https://git.example/team/service.git?token=secret", ""},
		{"url-fragment", "service", "https://git.example/team/service.git#main", ""},
		{"ssh-query", "service", "ssh://git@git.example/team/service.git?x=1", ""},
		{"file-url", "service", "file:///tmp/template.git", ""},
		{"bare-path", "service", "/tmp/template.git", ""},
		{"bad-module-path", "service", "https://git.example/team/../service.git", ""},
		{"option-ref", "service", "https://git.example/team/service.git", "--upload-pack=bad"},
		{"refspec", "service", "https://git.example/team/service.git", "main:other"},
		{"space-ref", "service", "https://git.example/team/service.git", "bad ref"},
		{"traversal-ref", "service", "https://git.example/team/service.git", "../main"},
		{"range-ref", "service", "https://git.example/team/service.git", "main..release"},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			root := t.TempDir()
			parent := filepath.Join(root, "not-created")
			writeFixtureFile(t, root, "sentinel.txt", "keep unrelated content\n")
			_, err := generate(context.Background(), Options{
				Name: tc.name, GitURL: tc.url, OutputDir: parent, TemplateRef: tc.ref,
			}, template)
			require.Error(t, err)
			assertPathAbsent(t, parent)
			assertPathAbsent(t, filepath.Join(root, "escaped"))
			assert.Equal(t, "keep unrelated content\n", readGeneratedFile(t, root, "sentinel.txt"))
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			require.Len(t, entries, 1, "invalid input must not create output or staging paths")
			assert.Equal(t, "sentinel.txt", entries[0].Name())
		})
	}
}

func TestGenerateRejectsTemplateSymlinksWithoutCopyingTheirTargets(t *testing.T) {
	isolateGit(t)
	for _, linkPath := range []string{"leaked.txt", "nested/link", ".env"} {
		t.Run(linkPath, func(t *testing.T) {
			template, _ := newTemplate(t, nil)
			secretDirectory := t.TempDir()
			writeFixtureFile(t, secretDirectory, "secret.txt", "external secret must not be copied\n")
			secretPath := filepath.Join(secretDirectory, "secret.txt")
			link := filepath.Join(template, filepath.FromSlash(linkPath))
			require.NoError(t, os.MkdirAll(filepath.Dir(link), 0o755))
			require.NoError(t, os.Symlink(secretPath, link))
			fixtureGit(t, template, "add", "--all")
			fixtureGit(t, template, "commit", "-m", "template symlink")
			parent := t.TempDir()
			writeFixtureFile(t, parent, "sentinel.txt", "keep sibling\n")

			_, err := generate(context.Background(), Options{
				Name: "unsafe", GitURL: "https://git.example/team/unsafe.git", OutputDir: parent, NoGit: true,
			}, template)
			require.Error(t, err)
			assertPathAbsent(t, filepath.Join(parent, "unsafe"))
			assert.Equal(t, "external secret must not be copied\n", readGeneratedFile(t, secretDirectory, "secret.txt"))
			assert.Equal(t, "keep sibling\n", readGeneratedFile(t, parent, "sentinel.txt"))
			entries, err := os.ReadDir(parent)
			require.NoError(t, err)
			require.Len(t, entries, 1, "rejected symlinks must not leave copied files or staging paths")
			assert.Equal(t, "sentinel.txt", entries[0].Name())
			info, err := os.Lstat(link)
			require.NoError(t, err)
			assert.NotZero(t, info.Mode()&os.ModeSymlink, "the source template must be unchanged")
		})
	}
}

// All Git subprocesses, including those launched by generate, inherit an
// isolated configuration. No user hooks, credentials, or global config apply.
func isolateGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git is not installed")
	}
	// Clear ambient repository/config overrides while letting t.Setenv restore
	// the caller's original environment after this test.
	for _, key := range []string{
		"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE",
		"GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES",
		"GIT_NAMESPACE", "GIT_CONFIG", "GIT_CONFIG_PARAMETERS",
	} {
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}
	home := t.TempDir()
	emptyHooks := filepath.Join(home, "empty-hooks")
	emptyTemplate := filepath.Join(home, "empty-template")
	require.NoError(t, os.Mkdir(emptyHooks, 0o755))
	require.NoError(t, os.Mkdir(emptyTemplate, 0o755))
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.hooksPath")
	t.Setenv("GIT_CONFIG_VALUE_0", emptyHooks)
	t.Setenv("GIT_TEMPLATE_DIR", emptyTemplate)
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	t.Setenv("GIT_ALLOW_PROTOCOL", "file")
	t.Setenv("GIT_AUTHOR_NAME", "Project Generator Fixture")
	t.Setenv("GIT_AUTHOR_EMAIL", "fixture@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "Project Generator Fixture")
	t.Setenv("GIT_COMMITTER_EMAIL", "fixture@example.invalid")
	t.Setenv("GIT_AUTHOR_DATE", "2025-01-01T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2025-01-01T00:00:00Z")
}

func newTemplate(t *testing.T, additionalFiles map[string]string) (string, string) {
	t.Helper()
	template := filepath.Join(t.TempDir(), "local template with spaces")
	require.NoError(t, os.Mkdir(template, 0o755))
	files := map[string]string{
		"go.mod":       "module " + fixtureModule + "\n\ngo 1.24.0\n",
		"go.sum":       "",
		"main.go":      "package main\n\nfunc main() {}\n",
		".env.example": "NAME=test\nJWT_ISSUER=hanif-skeleton\nDB_NAME=test_skeleton\n",
		"Makefile":     "BINARY_NAME=main\n\nbuild:\n\tgo build -o $(BINARY_NAME) .\n",
		"README.md":    "# Hanif Skeleton\n\nA small local template.\n",
	}
	for path, content := range additionalFiles {
		files[path] = content
	}
	for path, content := range files {
		writeFixtureFile(t, template, path, content)
	}
	fixtureGit(t, template, "init", "--initial-branch=main")
	fixtureGit(t, template, "add", "--all")
	fixtureGit(t, template, "-c", "commit.gpgSign=false", "commit", "-m", "initial template")
	return template, fixtureGit(t, template, "rev-parse", "HEAD")
}

func fixtureGit(t *testing.T, directory string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = directory
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v in %s: %s", args, directory, output)
	return strings.TrimSpace(string(output))
}

func writeFixtureFile(t *testing.T, directory, path, content string) {
	t.Helper()
	fullPath := filepath.Join(directory, filepath.FromSlash(path))
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o600))
}

func readGeneratedFile(t *testing.T, directory, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(path)))
	require.NoError(t, err)
	return string(content)
}

func assertPathAbsent(t *testing.T, path string) {
	t.Helper()
	_, err := os.Lstat(path)
	assert.True(t, os.IsNotExist(err), "expected %s not to exist; got %v", path, err)
}

func parseEnvironment(content string) map[string]string {
	values := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), "\"'")
		}
	}
	return values
}
