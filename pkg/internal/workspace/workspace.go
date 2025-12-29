// Package workspace manages the cf workspace
package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
	v1 "github.com/harshit-vibes/cf/pkg/internal/schema/v1"
	"gopkg.in/yaml.v3"
)

const (
	ManifestFile = "workspace.yaml"
	ProblemsDir  = "problems"
	TemplatesDir = "templates"
	SubmissionsDir = "submissions"
	StatsDir     = "stats"
)

// Workspace manages a cf workspace
type Workspace struct {
	root     string
	manifest *v1.Workspace
}

// New creates a new workspace manager
func New(root string) *Workspace {
	return &Workspace{root: root}
}

// Root returns the workspace root path
func (w *Workspace) Root() string {
	return w.root
}

// ManifestPath returns the path to workspace.yaml
func (w *Workspace) ManifestPath() string {
	return filepath.Join(w.root, ManifestFile)
}

// Exists checks if the workspace is initialized
func (w *Workspace) Exists() bool {
	_, err := os.Stat(w.ManifestPath())
	return err == nil
}

// Init initializes a new workspace
func (w *Workspace) Init(name, handle string) error {
	// Create root directory
	if err := os.MkdirAll(w.root, 0755); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	// Create manifest
	manifest := v1.NewWorkspace(name, handle)
	w.manifest = manifest

	// Save manifest
	if err := w.SaveManifest(); err != nil {
		return err
	}

	// Create subdirectories
	dirs := []string{
		filepath.Join(w.root, manifest.Paths.Problems),
		filepath.Join(w.root, manifest.Paths.Templates),
		filepath.Join(w.root, manifest.Paths.Submissions),
		filepath.Join(w.root, manifest.Paths.Stats),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	// Create .gitkeep files
	for _, dir := range dirs {
		gitkeep := filepath.Join(dir, ".gitkeep")
		if err := os.WriteFile(gitkeep, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to create .gitkeep: %w", err)
		}
	}

	return nil
}

// Load loads an existing workspace
func (w *Workspace) Load() error {
	manifest, err := w.LoadManifest()
	if err != nil {
		return err
	}
	w.manifest = manifest
	return nil
}

// LoadManifest loads the workspace manifest
func (w *Workspace) LoadManifest() (*v1.Workspace, error) {
	data, err := os.ReadFile(w.ManifestPath())
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest v1.Workspace
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// SaveManifest saves the workspace manifest
func (w *Workspace) SaveManifest() error {
	if w.manifest == nil {
		return fmt.Errorf("no manifest to save")
	}

	data, err := yaml.Marshal(w.manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(w.ManifestPath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// Manifest returns the current manifest
func (w *Workspace) Manifest() *v1.Workspace {
	return w.manifest
}

// GetSchemaVersion returns the schema version from manifest
func (w *Workspace) GetSchemaVersion() (schema.Version, error) {
	manifest, err := w.LoadManifest()
	if err != nil {
		return schema.Version{}, err
	}
	return schema.ParseVersion(manifest.Schema.Version)
}

// ProblemsPath returns the problems directory path
func (w *Workspace) ProblemsPath() string {
	if w.manifest != nil && w.manifest.Paths.Problems != "" {
		return filepath.Join(w.root, w.manifest.Paths.Problems)
	}
	return filepath.Join(w.root, ProblemsDir)
}

// ProblemPath returns the path for a specific problem
func (w *Workspace) ProblemPath(platform string, contestID int, index string) string {
	return filepath.Join(
		w.ProblemsPath(),
		platform,
		"contest",
		fmt.Sprintf("%d", contestID),
		index,
	)
}

// TemplatesPath returns the templates directory path
func (w *Workspace) TemplatesPath() string {
	if w.manifest != nil && w.manifest.Paths.Templates != "" {
		return filepath.Join(w.root, w.manifest.Paths.Templates)
	}
	return filepath.Join(w.root, TemplatesDir)
}

// SubmissionsPath returns the submissions directory path
func (w *Workspace) SubmissionsPath() string {
	if w.manifest != nil && w.manifest.Paths.Submissions != "" {
		return filepath.Join(w.root, w.manifest.Paths.Submissions)
	}
	return filepath.Join(w.root, SubmissionsDir)
}

// StatsPath returns the stats directory path
func (w *Workspace) StatsPath() string {
	if w.manifest != nil && w.manifest.Paths.Stats != "" {
		return filepath.Join(w.root, w.manifest.Paths.Stats)
	}
	return filepath.Join(w.root, StatsDir)
}

// Validate checks the workspace integrity
func (w *Workspace) Validate() error {
	// Check manifest exists
	if !w.Exists() {
		return fmt.Errorf("workspace not initialized")
	}

	// Load and validate manifest
	manifest, err := w.LoadManifest()
	if err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	// Validate schema version
	version, err := schema.ParseVersion(manifest.Schema.Version)
	if err != nil {
		return fmt.Errorf("invalid schema version: %w", err)
	}

	if !schema.CurrentVersion.IsCompatible(version) {
		return fmt.Errorf("incompatible schema version: %s (current: %s)",
			version, schema.CurrentVersion)
	}

	// Check required directories
	dirs := []string{
		w.ProblemsPath(),
		w.TemplatesPath(),
		w.SubmissionsPath(),
		w.StatsPath(),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("missing directory: %s", dir)
		}
	}

	return nil
}
