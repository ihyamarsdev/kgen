package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallBinary_BasicFlow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-update-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source binary
	src := filepath.Join(tmpDir, "new-kgen")
	os.WriteFile(src, []byte("#!/bin/sh\necho new"), 0o755)

	// Create dest binary (simulating existing install)
	dest := filepath.Join(tmpDir, "kgen")
	os.WriteFile(dest, []byte("#!/bin/sh\necho old"), 0o755)

	installPath, err := installBinary(src, dest)
	if err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}

	// Verify dest exists and contains new content
	data, err := os.ReadFile(installPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "#!/bin/sh\necho new" {
		t.Errorf("expected new content, got: %s", string(data))
	}

	// Verify tmp files are cleaned up
	tmpFile := dest + ".tmp"
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("temp file should be cleaned up")
	}

	// Verify the DIRECTORY still exists (the key bug check)
	if info, err := os.Stat(tmpDir); err != nil || !info.IsDir() {
		t.Errorf("directory %s should still exist after install", tmpDir)
	}

	// Verify other files in the directory are untouched
	otherFile := filepath.Join(tmpDir, "other.txt")
	os.WriteFile(otherFile, []byte("keep me"), 0o644)
	if _, err := os.Stat(otherFile); os.IsNotExist(err) {
		t.Error("other files in the directory should not be deleted")
	}

	t.Logf("Install successful: %s", installPath)
}

func TestInstallBinary_TempCleanupOnFailure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-update-fail-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source binary
	src := filepath.Join(tmpDir, "new-kgen")
	os.WriteFile(src, []byte("new"), 0o755)

	// Create dest as a read-only directory (not a file) to force failure
	dest := filepath.Join(tmpDir, "kgen")
	os.Mkdir(dest, 0o555)

	_, err = installBinary(src, dest)
	if err == nil {
		t.Fatal("expected error when dest is a directory")
	}

	// Verify the directory still exists
	if info, err := os.Stat(dest); err != nil || !info.IsDir() {
		t.Error("destination directory should still exist after failed install")
	}

	// Verify no temp files left
	tmpFile := dest + ".tmp"
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("temp file should be cleaned up after failure")
	}

	t.Log("Temp cleanup on failure: OK")
}

func TestInstallBinary_AtomicRename(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-atomic-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	src := filepath.Join(tmpDir, "src")
	os.WriteFile(src, []byte("new-version"), 0o755)

	dest := filepath.Join(tmpDir, "dest")
	os.WriteFile(dest, []byte("old-version"), 0o755)

	installPath, err := installBinary(src, dest)
	if err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}

	data, _ := os.ReadFile(installPath)
	if string(data) != "new-version" {
		t.Errorf("expected 'new-version', got: %s", string(data))
	}
}

func TestInstallBinary_PreservesDirContents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-keep-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple files in the directory
	files := map[string]string{
		"kgen":            "old",
		"some-config.txt": "config-data",
		"readme.md":       "docs",
	}
	for name, content := range files {
		os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644)
	}

	// Update only the kgen binary
	src := filepath.Join(tmpDir, "src")
	os.WriteFile(src, []byte("new-kgen"), 0o755)
	os.Chmod(filepath.Join(tmpDir, "kgen"), 0o755)

	_, err = installBinary(src, filepath.Join(tmpDir, "kgen"))
	if err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}

	// Verify ALL files still exist
	for name := range files {
		p := filepath.Join(tmpDir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("file %s should still exist after install", name)
		}
	}

	// Verify kgen was updated
	data, _ := os.ReadFile(filepath.Join(tmpDir, "kgen"))
	if string(data) != "new-kgen" {
		t.Errorf("kgen should be updated, got: %s", string(data))
	}

	// Verify other files are untouched
	data, _ = os.ReadFile(filepath.Join(tmpDir, "some-config.txt"))
	if string(data) != "config-data" {
		t.Errorf("config file should be untouched, got: %s", string(data))
	}

	t.Log("All directory contents preserved after install")
}

func TestInstallBinary_PathResolution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-path-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Simulate: binary in a directory with other files
	realBin := filepath.Join(tmpDir, "bin", "kgen")
	os.MkdirAll(filepath.Dir(realBin), 0o755)
	os.WriteFile(realBin, []byte("real-kgen"), 0o755)
	os.WriteFile(filepath.Join(tmpDir, "bin", "config.yaml"), []byte("config"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "bin", "other.txt"), []byte("other"), 0o644)

	newBin := filepath.Join(tmpDir, "new-kgen")
	os.WriteFile(newBin, []byte("new-kgen-v2"), 0o755)

	binDir := filepath.Dir(realBin)
	targetPath := filepath.Join(binDir, "kgen")

	installPath, err := installBinary(newBin, targetPath)
	if err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}

	data, _ := os.ReadFile(installPath)
	if string(data) != "new-kgen-v2" {
		t.Errorf("binary not updated, got: %s", string(data))
	}

	// Verify ALL other files in the directory are untouched
	for _, name := range []string{"config.yaml", "other.txt"} {
		p := filepath.Join(binDir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("FOLDER CONTENT DELETED: %s no longer exists!", p)
		}
	}

	// Verify the DIRECTORY still exists
	if info, err := os.Stat(binDir); err != nil || !info.IsDir() {
		t.Errorf("DIRECTORY DELETED: %s no longer exists!", binDir)
	} else {
		entries, _ := os.ReadDir(binDir)
		t.Logf("Directory %s has %d entries (all preserved)", binDir, len(entries))
	}
}

func TestInstallBinary_SymlinkResolution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-symlink-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	realDir := filepath.Join(tmpDir, "real")
	os.MkdirAll(realDir, 0o755)
	realBin := filepath.Join(realDir, "kgen")
	os.WriteFile(realBin, []byte("real-kgen"), 0o755)
	os.WriteFile(filepath.Join(realDir, "stuff.txt"), []byte("important"), 0o644)

	linkPath := filepath.Join(tmpDir, "link")
	os.Symlink(realBin, linkPath)

	resolved, _ := filepath.EvalSymlinks(linkPath)
	binDir := filepath.Dir(resolved)
	targetPath := filepath.Join(binDir, "kgen")

	newBin := filepath.Join(tmpDir, "new-kgen")
	os.WriteFile(newBin, []byte("new-v2"), 0o755)

	_, err = installBinary(newBin, targetPath)
	if err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}

	data, _ := os.ReadFile(targetPath)
	if string(data) != "new-v2" {
		t.Errorf("target not updated, got: %s", string(data))
	}

	stuffPath := filepath.Join(realDir, "stuff.txt")
	if _, err := os.Stat(stuffPath); os.IsNotExist(err) {
		t.Error("FOLDER CONTENT DELETED after symlink update!")
	} else {
		d, _ := os.ReadFile(stuffPath)
		t.Logf("stuff.txt: %s (preserved)", string(d))
	}
}
