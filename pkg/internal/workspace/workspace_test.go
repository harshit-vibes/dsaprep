package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
)

func TestNew(t *testing.T) {
	ws := New("/tmp/test-workspace")

	if ws == nil {
		t.Fatal("New() returned nil")
	}
	if ws.Root() != "/tmp/test-workspace" {
		t.Errorf("Root() = %v, want %v", ws.Root(), "/tmp/test-workspace")
	}
}

func TestWorkspace_ManifestPath(t *testing.T) {
	ws := New("/tmp/test-workspace")
	want := "/tmp/test-workspace/workspace.yaml"

	if got := ws.ManifestPath(); got != want {
		t.Errorf("ManifestPath() = %v, want %v", got, want)
	}
}

func TestWorkspace_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Should not exist before init
	if ws.Exists() {
		t.Error("Exists() should return false before Init()")
	}

	// Initialize workspace
	err := ws.Init("Test Workspace", "testuser")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Should exist after init
	if !ws.Exists() {
		t.Error("Exists() should return true after Init()")
	}
}

func TestWorkspace_Init(t *testing.T) {
	tmpDir := t.TempDir()
	wsRoot := filepath.Join(tmpDir, "my-workspace")
	ws := New(wsRoot)

	err := ws.Init("DSA Practice", "tourist")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Check workspace.yaml exists
	if _, err := os.Stat(ws.ManifestPath()); os.IsNotExist(err) {
		t.Error("Init() did not create workspace.yaml")
	}

	// Check subdirectories exist
	dirs := []string{
		filepath.Join(wsRoot, "problems"),
		filepath.Join(wsRoot, "templates"),
		filepath.Join(wsRoot, "submissions"),
		filepath.Join(wsRoot, "stats"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Init() did not create directory %s", dir)
		}

		// Check .gitkeep exists
		gitkeep := filepath.Join(dir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			t.Errorf("Init() did not create .gitkeep in %s", dir)
		}
	}

	// Check manifest content
	manifest := ws.Manifest()
	if manifest == nil {
		t.Fatal("Manifest() returned nil after Init()")
	}
	if manifest.Name != "DSA Practice" {
		t.Errorf("Manifest().Name = %v, want %v", manifest.Name, "DSA Practice")
	}
	if manifest.Codeforces.Handle != "tourist" {
		t.Errorf("Manifest().Codeforces.Handle = %v, want %v", manifest.Codeforces.Handle, "tourist")
	}
}

func TestWorkspace_Load(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Initialize first
	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Create new workspace instance and load
	ws2 := New(tmpDir)
	err = ws2.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if ws2.Manifest() == nil {
		t.Error("Load() did not populate manifest")
	}
	if ws2.Manifest().Name != "Test" {
		t.Errorf("Loaded manifest name = %v, want %v", ws2.Manifest().Name, "Test")
	}
}

func TestWorkspace_LoadManifest_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	_, err := ws.LoadManifest()
	if err == nil {
		t.Error("LoadManifest() should return error for non-existent file")
	}
}

func TestWorkspace_SaveManifest_NoManifest(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.SaveManifest()
	if err == nil {
		t.Error("SaveManifest() should return error when manifest is nil")
	}
}

func TestWorkspace_GetSchemaVersion(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	version, err := ws.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion() error = %v", err)
	}

	if version != schema.CurrentVersion {
		t.Errorf("GetSchemaVersion() = %v, want %v", version, schema.CurrentVersion)
	}
}

func TestWorkspace_ProblemsPath(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(ws *Workspace)
		wantSuffix string
	}{
		{
			name:       "default path without manifest",
			setupFunc:  func(ws *Workspace) {},
			wantSuffix: "problems",
		},
		{
			name: "path from manifest",
			setupFunc: func(ws *Workspace) {
				ws.Init("Test", "user")
			},
			wantSuffix: "problems",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			ws := New(tmpDir)
			tt.setupFunc(ws)

			got := ws.ProblemsPath()
			want := filepath.Join(tmpDir, tt.wantSuffix)
			if got != want {
				t.Errorf("ProblemsPath() = %v, want %v", got, want)
			}
		})
	}
}

func TestWorkspace_ProblemPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	got := ws.ProblemPath("codeforces", 1325, "A")
	want := filepath.Join(tmpDir, "problems", "codeforces", "contest", "1325", "A")

	if got != want {
		t.Errorf("ProblemPath() = %v, want %v", got, want)
	}
}

func TestWorkspace_TemplatesPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	got := ws.TemplatesPath()
	want := filepath.Join(tmpDir, "templates")

	if got != want {
		t.Errorf("TemplatesPath() = %v, want %v", got, want)
	}
}

func TestWorkspace_SubmissionsPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	got := ws.SubmissionsPath()
	want := filepath.Join(tmpDir, "submissions")

	if got != want {
		t.Errorf("SubmissionsPath() = %v, want %v", got, want)
	}
}

func TestWorkspace_StatsPath(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	got := ws.StatsPath()
	want := filepath.Join(tmpDir, "stats")

	if got != want {
		t.Errorf("StatsPath() = %v, want %v", got, want)
	}
}

func TestWorkspace_Validate(t *testing.T) {
	t.Run("not initialized", func(t *testing.T) {
		tmpDir := t.TempDir()
		ws := New(tmpDir)

		err := ws.Validate()
		if err == nil {
			t.Error("Validate() should return error for uninitialized workspace")
		}
	})

	t.Run("valid workspace", func(t *testing.T) {
		tmpDir := t.TempDir()
		ws := New(tmpDir)

		err := ws.Init("Test", "user")
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		err = ws.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("missing directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		ws := New(tmpDir)

		err := ws.Init("Test", "user")
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		// Remove a required directory
		os.RemoveAll(filepath.Join(tmpDir, "problems"))

		err = ws.Validate()
		if err == nil {
			t.Error("Validate() should return error when directory is missing")
		}
	})
}

func TestConstants(t *testing.T) {
	if ManifestFile != "workspace.yaml" {
		t.Errorf("ManifestFile = %v, want workspace.yaml", ManifestFile)
	}
	if ProblemsDir != "problems" {
		t.Errorf("ProblemsDir = %v, want problems", ProblemsDir)
	}
	if TemplatesDir != "templates" {
		t.Errorf("TemplatesDir = %v, want templates", TemplatesDir)
	}
	if SubmissionsDir != "submissions" {
		t.Errorf("SubmissionsDir = %v, want submissions", SubmissionsDir)
	}
	if StatsDir != "stats" {
		t.Errorf("StatsDir = %v, want stats", StatsDir)
	}
}
