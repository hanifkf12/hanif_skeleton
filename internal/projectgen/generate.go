// Package projectgen creates an independent project from the skeleton template.
package projectgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	DefaultTemplateURL = "https://github.com/hanifkf12/hanif_skeleton.git"
	DefaultTemplateRef = "main"
)

type Options struct {
	Name        string
	GitURL      string
	OutputDir   string
	TemplateRef string
	NoGit       bool
}

type Result struct {
	Directory      string
	ModulePath     string
	TemplateCommit string
}

// Generate copies the selected template revision into a new project directory.
// GitURL names the new repository; generation never contacts that repository.
func Generate(ctx context.Context, options Options) (Result, error) {
	return generate(ctx, options, DefaultTemplateURL)
}

func generate(ctx context.Context, options Options, templateURL string) (result Result, err error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if err := validateName(options.Name); err != nil {
		return Result{}, err
	}
	modulePath, err := repositoryModule(options.GitURL)
	if err != nil {
		return Result{}, err
	}
	if options.TemplateRef == "" {
		options.TemplateRef = DefaultTemplateRef
	}
	if err := validateRef(options.TemplateRef); err != nil {
		return Result{}, err
	}
	if templateURL == "" || strings.HasPrefix(templateURL, "-") {
		return Result{}, fmt.Errorf("invalid template repository")
	}
	if options.OutputDir == "" {
		options.OutputDir = "."
	}
	parent, err := filepath.Abs(options.OutputDir)
	if err != nil {
		return Result{}, fmt.Errorf("resolve output parent: %w", err)
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output parent: %w", err)
	}
	directory := filepath.Join(parent, options.Name)
	// Mkdir, rather than a check followed by MkdirAll, exclusively claims the target.
	if err := os.Mkdir(directory, 0o755); err != nil {
		return Result{}, fmt.Errorf("create project directory (target must not exist): %w", err)
	}
	owned, err := os.Lstat(directory)
	if err != nil {
		return Result{}, fmt.Errorf("identify project directory: %w", err)
	}
	defer func() {
		if err == nil {
			return
		}
		result = Result{}
		current, statErr := os.Lstat(directory)
		if errors.Is(statErr, os.ErrNotExist) {
			return
		}
		if statErr != nil {
			err = errors.Join(err, fmt.Errorf("identify failed output for cleanup: %w", statErr))
			return
		}
		if !os.SameFile(owned, current) {
			err = errors.Join(err, fmt.Errorf("output directory changed ownership; refusing cleanup"))
			return
		}
		if cleanupErr := os.RemoveAll(directory); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("clean failed output: %w", cleanupErr))
		}
	}()

	template, err := os.MkdirTemp("", "skeleton-template-")
	if err != nil {
		return Result{}, fmt.Errorf("create temporary template directory: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(template); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("clean temporary template: %w", cleanupErr))
		}
	}()
	if _, err = runGit(ctx, template, "init", "--initial-branch=main", "--template=", template); err != nil {
		return Result{}, err
	}
	if _, err = runGit(ctx, template, "fetch", "--depth=1", "--no-tags", "--", templateURL, options.TemplateRef); err != nil {
		return Result{}, err
	}
	if _, err = runGit(ctx, template, "checkout", "--detach", "--force", "FETCH_HEAD"); err != nil {
		return Result{}, err
	}
	commit, err := runGit(ctx, template, "rev-parse", "--verify", "HEAD")
	if err != nil {
		return Result{}, err
	}
	if err = copyTemplate(ctx, template, directory); err != nil {
		return Result{}, err
	}
	if err = rewriteProject(ctx, directory, options, modulePath); err != nil {
		return Result{}, err
	}
	if !options.NoGit {
		if _, err = runGit(ctx, directory, "init", "--initial-branch=main", "--template=", directory); err != nil {
			return Result{}, err
		}
		if _, err = runGit(ctx, directory, "remote", "add", "origin", options.GitURL); err != nil {
			return Result{}, err
		}
	}
	if err = ctx.Err(); err != nil {
		return Result{}, err
	}
	return Result{Directory: directory, ModulePath: modulePath, TemplateCommit: strings.TrimSpace(commit)}, nil
}

func runGit(ctx context.Context, directory string, args ...string) (string, error) {
	// Explicit paths and configuration prevent ambient worktree and symlink
	// settings from redirecting writes or hiding tracked symbolic links.
	prefix := []string{
		"--git-dir=" + filepath.Join(directory, ".git"),
		"--work-tree=" + directory,
		"-c", "core.symlinks=true",
		"-c", "core.autocrlf=false",
		"-c", "core.hooksPath=" + os.DevNull,
	}
	command := exec.CommandContext(ctx, "git", append(prefix, args...)...)
	command.Dir = directory
	environment := os.Environ()
	command.Env = make([]string, 0, len(environment)+1)
	for _, variable := range environment {
		key, _, _ := strings.Cut(variable, "=")
		switch key {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE",
			"GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_NAMESPACE":
			continue
		}
		command.Env = append(command.Env, variable)
	}
	command.Env = append(command.Env, "GIT_TERMINAL_PROMPT=0")
	var diagnostics bytes.Buffer
	command.Stderr = &diagnostics
	output, err := command.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("git %s: %w", args[0], ctx.Err())
		}
		return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(diagnostics.String()))
	}
	return string(output), nil
}

func copyTemplate(ctx context.Context, source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("read template: %w", walkErr)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if relative == ".git" {
			return filepath.SkipDir
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("template contains symbolic link %q", relative)
		}
		// Do not skip excluded directories: their contents must still be checked
		// for symlinks before the template is accepted.
		if excludedTemplatePath(filepath.ToSlash(relative)) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect template file %q: %w", relative, err)
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			if err := os.Mkdir(target, 0o755); err != nil {
				return fmt.Errorf("create template directory %q: %w", relative, err)
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("template contains unsupported file %q", relative)
		}
		if err := copyRegularFile(path, target, info.Mode().Perm()); err != nil {
			return fmt.Errorf("copy template file %q: %w", relative, err)
		}
		return nil
	})
}

func excludedTemplatePath(path string) bool {
	if path == "cmd/skeleton" || strings.HasPrefix(path, "cmd/skeleton/") || path == "internal/projectgen" || strings.HasPrefix(path, "internal/projectgen/") {
		return true
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".git" || part == ".idea" || part == ".vscode" || part == ".zed" || part == "bin" || part == ".env" || strings.HasPrefix(part, ".env.") && part != ".env.example" || strings.HasPrefix(part, "coverage.") || strings.HasSuffix(part, ".out") || strings.HasSuffix(part, ".test") {
			return true
		}
	}
	return false
}

func copyRegularFile(source, destination string, mode fs.FileMode) (err error) {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, input.Close()) }()
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, output.Close()) }()
	if _, err = io.Copy(output, input); err != nil {
		return err
	}
	return output.Chmod(mode)
}
