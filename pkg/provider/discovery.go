// SPDX-License-Identifier: EUPL-1.2

package provider

import (
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

	files := providerManifestFiles(dir)
	if len(files) == 0 {
		return nil, nil
	}

	providers := make([]Provider, 0, len(files))
	for _, file := range files {
		provider, err := loadProviderManifest(file)
		if err != nil {
			return nil, core.E(op, core.Sprintf("load %s", file), err)
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

func providerManifestFiles(dir string) []string {
	dir = core.Trim(dir)
	if dir == "" {
		dir = DefaultProvidersDir
	}

	files := append(core.PathGlob(core.JoinPath(dir, "*.yaml")), core.PathGlob(core.JoinPath(dir, "*.yml"))...)
	slices.Sort(files)
	return files
}

func loadProviderManifest(path string) (Provider, error) {
	const op = "provider.loadProviderManifest"

	result := (&core.Fs{}).New("/").Read(path)
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

	cfg, err := manifest.proxyConfig(path)
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
