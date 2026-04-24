package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONInputArg_Inline(t *testing.T) {
	data, err := ReadJSONInputArg(`{"name":"test"}`)
	if err != nil {
		t.Fatalf("ReadJSONInputArg inline: %v", err)
	}
	if string(data) != `{"name":"test"}` {
		t.Fatalf("got %q", string(data))
	}
}

func TestReadJSONInputArg_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.json")
	if err := os.WriteFile(path, []byte(`{"from":"file"}`), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	data, err := ReadJSONInputArg("@" + path)
	if err != nil {
		t.Fatalf("ReadJSONInputArg file: %v", err)
	}
	if string(data) != `{"from":"file"}` {
		t.Fatalf("got %q", string(data))
	}
}

func TestIsStdinJSONInputArg(t *testing.T) {
	if !IsStdinJSONInputArg("@-") {
		t.Fatal("expected @- to be stdin")
	}
	if !IsStdinJSONInputArg("-") {
		t.Fatal("expected - to be accepted as stdin")
	}
}
