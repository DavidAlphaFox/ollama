package server

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type ModelPath struct {
	ProtocolScheme string
	Registry       string
	Namespace      string
	Repository     string
	Tag            string
}

const (
	DefaultRegistry       = "registry.ollama.ai"
	DefaultNamespace      = "library"
	DefaultTag            = "latest"
	DefaultProtocolScheme = "https"
)

var (
	ErrInvalidImageFormat = errors.New("invalid image format")
	ErrInvalidProtocol    = errors.New("invalid protocol scheme")
	ErrInsecureProtocol   = errors.New("insecure protocol http")
)

func ParseModelPath(name string) (*ModelPath, error) {
	mp := ModelPath{
		ProtocolScheme: DefaultProtocolScheme,
		Registry:       DefaultRegistry,
		Namespace:      DefaultNamespace,
		Repository:     "",
		Tag:            DefaultTag,
	}

	before, after, found := strings.Cut(name, "://")
	if found {
		mp.ProtocolScheme = before
		name = after
	}

	parts := strings.Split(name, string(os.PathSeparator))
	switch len(parts) {
	case 3:
		mp.Registry = parts[0]
		mp.Namespace = parts[1]
		mp.Repository = parts[2]
	case 2:
		mp.Namespace = parts[0]
		mp.Repository = parts[1]
	case 1:
		mp.Repository = parts[0]
	}

	if repo, tag, found := strings.Cut(mp.Repository, ":"); found {
		mp.Repository = repo
		mp.Tag = tag
	}

	if mp.Repository == "" {
		return nil, os.ErrNotExist
	}

	return &mp, nil
}

func (mp ModelPath) GetNamespaceRepository() string {
	return fmt.Sprintf("%s/%s", mp.Namespace, mp.Repository)
}

func (mp ModelPath) GetFullTagname() string {
	return fmt.Sprintf("%s/%s/%s:%s", mp.Registry, mp.Namespace, mp.Repository, mp.Tag)
}

func (mp ModelPath) GetShortTagname() string {
	if mp.Registry == DefaultRegistry {
		if mp.Namespace == DefaultNamespace {
			return fmt.Sprintf("%s:%s", mp.Repository, mp.Tag)
		}
		return fmt.Sprintf("%s/%s:%s", mp.Namespace, mp.Repository, mp.Tag)
	}
	return fmt.Sprintf("%s/%s/%s:%s", mp.Registry, mp.Namespace, mp.Repository, mp.Tag)
}

func (mp ModelPath) GetManifestPath(createDir bool) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".ollama", "models", "manifests", mp.Registry, mp.Namespace, mp.Repository, mp.Tag)
	if createDir {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}
	}

	return path, nil
}

func (mp ModelPath) BaseURL() *url.URL {
	return &url.URL{
		Scheme: mp.ProtocolScheme,
		Host:   mp.Registry,
	}
}

func GetManifestPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".ollama", "models", "manifests")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", err
	}

	return path, nil
}

func GetBlobsPath(digest string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		digest = strings.ReplaceAll(digest, ":", "-")
	}

	path := filepath.Join(home, ".ollama", "models", "blobs", digest)
	dirPath := filepath.Dir(path)
	if digest == "" {
		dirPath = path
	}

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return "", err
	}

	return path, nil
}
