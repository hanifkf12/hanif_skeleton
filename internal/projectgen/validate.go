package projectgen

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"golang.org/x/mod/module"
)

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("project name is required")
	}
	for i := range len(name) {
		c := name[i]
		alphanumeric := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
		if !alphanumeric && (i == 0 || c != '_' && c != '-') {
			return fmt.Errorf("project name must begin with an ASCII letter or digit and contain only ASCII letters, digits, underscores, or hyphens")
		}
	}
	return nil
}

func repositoryModule(raw string) (string, error) {
	invalid := func() (string, error) {
		return "", fmt.Errorf("repository URL must be an HTTP(S), ssh://git@host/group/repo.git, or git@host:group/repo.git hosted repository URL without credentials, query, or fragment")
	}
	if raw == "" || strings.HasPrefix(raw, "-") || strings.ContainsAny(raw, "?#\\") || strings.IndexFunc(raw, unicode.IsSpace) >= 0 || strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return invalid()
	}
	var host, path string
	if strings.HasPrefix(raw, "git@") {
		host, path, _ = strings.Cut(strings.TrimPrefix(raw, "git@"), ":")
		if host == "" || path == "" || strings.ContainsAny(host, "@/:") || strings.HasPrefix(path, "/") {
			return invalid()
		}
	} else {
		u, err := url.Parse(raw)
		if err != nil || u.Opaque != "" || u.Host == "" || u.RawQuery != "" || u.Fragment != "" || u.ForceQuery || u.RawPath != "" || strings.Contains(raw, "%") {
			return invalid()
		}
		switch u.Scheme {
		case "http", "https":
			if u.User != nil {
				return invalid()
			}
		case "ssh":
			if u.User == nil || u.User.Username() != "git" {
				return invalid()
			}
			if _, password := u.User.Password(); password {
				return invalid()
			}
		default:
			return invalid()
		}
		host, path = u.Hostname(), strings.TrimPrefix(u.Path, "/")
		if u.Port() != "" {
			for _, c := range u.Port() {
				if c < '0' || c > '9' {
					return invalid()
				}
			}
		}
	}
	// A module path represents the hosted repository, not its transport port.
	if host == "" || strings.ContainsAny(host, ":[]") || path == "" || strings.HasSuffix(path, "/") {
		return invalid()
	}
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return invalid()
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.HasPrefix(part, "-") {
			return invalid()
		}
	}
	path = strings.TrimSuffix(path, ".git")
	modulePath := strings.ToLower(host) + "/" + path
	if err := module.CheckPath(modulePath); err != nil {
		return "", fmt.Errorf("invalid repository module path %q: %w", modulePath, err)
	}
	return modulePath, nil
}

func validateRef(ref string) error {
	// Match Git's check-ref-format rules while also rejecting refspec syntax.
	if ref == "" || ref == "@" || strings.HasPrefix(ref, "-") || strings.HasPrefix(ref, "+") || strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") || strings.HasSuffix(ref, ".") || strings.Contains(ref, "..") || strings.Contains(ref, "@{") || strings.ContainsAny(ref, " ~^:?*[\\") {
		return fmt.Errorf("invalid template ref %q", ref)
	}
	for _, c := range ref {
		if c < 0x20 || c == 0x7f {
			return fmt.Errorf("invalid template ref %q", ref)
		}
	}
	for _, part := range strings.Split(ref, "/") {
		if part == "" || strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".lock") {
			return fmt.Errorf("invalid template ref %q", ref)
		}
	}
	return nil
}
