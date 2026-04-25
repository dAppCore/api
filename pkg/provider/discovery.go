// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"os"
	"path/filepath"
	"slices"

	core "dappco.re/go/core"
	"gopkg.in/yaml.v3"
)

const DefaultProvidersDir = ".core/providers"

// Discover loads local polyglot provider manifests from dir. A blank dir uses
// ".core/providers". Missing directories or no matching YAML files are treated
// as empty discovery results.
func Discover(dir string) ([]Provider, error) {
	const op = "provider.Discover"

	canonicalDir, files, err := providerManifestFiles(dir)
	if err != nil {
		return nil, core.E(op, "discover manifest files", err)
	}
	if len(files) == 0 {
		return nil, nil
	}

	fs := (&core.Fs{}).New(canonicalDir)
	providers := make([]Provider, 0, len(files))
	for _, file := range files {
		provider, err := loadProviderManifest(fs, file)
		if err != nil {
			return nil, core.E(op, core.Sprintf("load %s", file.path), err)
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

// DiscoverDefault loads provider manifests from ".core/providers".
func DiscoverDefault() ([]Provider, error) {
	return Discover(DefaultProvidersDir)
}

// Discover adds every provider manifest found in dir to the registry.
func (r *Registry) Discover(dir string) error {
	providers, err := Discover(dir)
	if err != nil {
		return err
	}
	for _, p := range providers {
		r.Add(p)
	}
	return nil
}

// DiscoverDefault adds providers from ".core/providers" to the registry.
func (r *Registry) DiscoverDefault() error {
	return r.Discover(DefaultProvidersDir)
}

type providerManifest struct {
	Name          string                  `yaml:"name"`
	Runtime       string                  `yaml:"runtime"`
	Kind          string                  `yaml:"kind"`
	Type          string                  `yaml:"type"`
	BasePath      string                  `yaml:"basePath"`
	BasePathSnake string                  `yaml:"base_path"`
	Upstream      string                  `yaml:"upstream"`
	SpecFile      string                  `yaml:"specFile"`
	SpecFileSnake string                  `yaml:"spec_file"`
	Element       providerElementManifest `yaml:"element"`
}

type providerElementManifest struct {
	Tag    string `yaml:"tag"`
	Source string `yaml:"source"`
}

type providerManifestFile struct {
	path     string
	readPath string
}

func providerManifestFiles(dir string) (string, []providerManifestFile, error) {
	dir = core.Trim(dir)
	if dir == "" {
		dir = DefaultProvidersDir
	}

	canonicalDir, ok, err := canonicalProviderDir(dir)
	if err != nil || !ok {
		return canonicalDir, nil, err
	}

	matches := append(core.PathGlob(filepath.Join(canonicalDir, "*.yaml")), core.PathGlob(filepath.Join(canonicalDir, "*.yml"))...)
	slices.Sort(matches)

	files := make([]providerManifestFile, 0, len(matches))
	for _, match := range matches {
		file, err := canonicalProviderManifestFile(canonicalDir, match)
		if err != nil {
			return "", nil, err
		}
		files = append(files, file)
	}
	return canonicalDir, files, nil
}

func canonicalProviderDir(dir string) (string, bool, error) {
	const op = "provider.providerManifestFiles"

	absolute, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return "", false, core.E(op, "resolve provider directory path", err)
	}

	info, err := os.Lstat(absolute)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, core.E(op, "stat provider directory", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", false, core.E(op, "symlinked provider directory rejected: "+absolute, nil)
	}
	if !info.IsDir() {
		return "", false, core.E(op, "provider path is not a directory: "+absolute, nil)
	}

	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", false, core.E(op, "resolve provider directory symlinks: "+absolute, err)
	}
	return filepath.Clean(resolved), true, nil
}

func canonicalProviderManifestFile(canonicalDir, path string) (providerManifestFile, error) {
	const op = "provider.providerManifestFiles"

	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return providerManifestFile{}, core.E(op, "resolve provider manifest path", err)
	}

	info, err := os.Lstat(absolute)
	if err != nil {
		return providerManifestFile{}, core.E(op, "stat provider manifest", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return providerManifestFile{}, core.E(op, "symlinked provider manifest rejected: "+absolute, nil)
	}
	if !info.Mode().IsRegular() {
		return providerManifestFile{}, core.E(op, "provider manifest is not a regular file: "+absolute, nil)
	}

	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return providerManifestFile{}, core.E(op, "resolve provider manifest symlinks: "+absolute, err)
	}
	resolved = filepath.Clean(resolved)

	relative, err := filepath.Rel(canonicalDir, resolved)
	if err != nil {
		return providerManifestFile{}, core.E(op, "compare provider manifest with provider directory", err)
	}
	parentPrefix := ".." + string(filepath.Separator)
	if relative == ".." || core.HasPrefix(relative, parentPrefix) || filepath.IsAbs(relative) {
		return providerManifestFile{}, core.E(op, "provider manifest escapes provider directory: "+absolute, nil)
	}

	readPath := relative
	if canonicalDir == string(filepath.Separator) {
		readPath = resolved
	}

	return providerManifestFile{
		path:     resolved,
		readPath: readPath,
	}, nil
}

func loadProviderManifest(fs *core.Fs, file providerManifestFile) (Provider, error) {
	const op = "provider.loadProviderManifest"

	result := fs.Read(file.readPath)
	if !result.OK {
		if err, ok := result.Value.(error); ok {
			return nil, core.E(op, "read manifest", err)
		}
		return nil, core.E(op, "read manifest", nil)
	}

	raw, ok := result.Value.(string)
	if !ok {
		return nil, core.E(op, "manifest content was not text", nil)
	}

	var manifest providerManifest
	if err := yaml.Unmarshal([]byte(raw), &manifest); err != nil {
		return nil, core.E(op, "parse manifest yaml", err)
	}

	cfg, err := manifest.proxyConfig(file.path)
	if err != nil {
		return nil, err
	}

	p := NewProxy(cfg)
	if err := p.Err(); err != nil {
		return nil, err
	}
	return p, nil
}

func (m providerManifest) proxyConfig(path string) (ProxyConfig, error) {
	const op = "provider.Manifest.proxyConfig"

	name := core.Trim(m.Name)
	if name == "" {
		return ProxyConfig{}, core.E(op, "name is required", nil)
	}

	basePath, err := normaliseManifestBasePath(firstNonEmpty(m.BasePath, m.BasePathSnake))
	if err != nil {
		return ProxyConfig{}, err
	}

	upstream := core.Trim(m.Upstream)
	if upstream == "" {
		return ProxyConfig{}, core.E(op, "upstream is required", nil)
	}

	return ProxyConfig{
		Name:     name,
		BasePath: basePath,
		Upstream: upstream,
		Element: ElementSpec{
			Tag:    core.Trim(m.Element.Tag),
			Source: core.Trim(m.Element.Source),
		},
		SpecFile: resolveManifestPath(path, firstNonEmpty(m.SpecFile, m.SpecFileSnake)),
	}, nil
}

func normaliseManifestBasePath(path string) (string, error) {
	path = core.Trim(path)
	if path == "" {
		return "", core.E("provider.normaliseManifestBasePath", "basePath is required", nil)
	}
	if !core.HasPrefix(path, "/") {
		path = "/" + path
	}
	for core.HasSuffix(path, "/") && path != "/" {
		path = core.TrimSuffix(path, "/")
	}
	return path, nil
}

func resolveManifestPath(manifestFile, value string) string {
	value = core.Trim(value)
	if value == "" || core.PathIsAbs(value) || core.HasPrefix(value, "http://") || core.HasPrefix(value, "https://") {
		return value
	}
	return core.CleanPath(core.JoinPath(core.PathDir(manifestFile), value), "/")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = core.Trim(value); value != "" {
			return value
		}
	}
	return ""
}
