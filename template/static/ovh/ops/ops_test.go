package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPreambleSurvivesShellQuoting(t *testing.T) {
	vars := scriptVars{
		OpsUser:    "ops",
		PublicKey:  "ssh-ed25519 AAAA it's $(rm -rf /) `id` \"quoted\"",
		HostKey:    "-----BEGIN OPENSSH PRIVATE KEY-----\nline'one\nline two\n-----END OPENSSH PRIVATE KEY-----\n",
		HostKeyPub: "ssh-ed25519 BBBB",
	}
	out, err := exec.Command("bash", "-c", vars.preamble()+
		`printf '%s\0%s\0%s\0%s' "$OPS_USER" "$PUBLIC_KEY" "$HOST_KEY" "$HOST_KEY_PUB"`).Output()
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(string(out), "\x00")
	want := []string{vars.OpsUser, vars.PublicKey, vars.HostKey, vars.HostKeyPub}
	if !slices.Equal(got, want) {
		t.Fatalf("bash read the preamble as %q, want %q", got, want)
	}
}

func TestHostKeyIsAnOpenSSHKeyPair(t *testing.T) {
	host, err := newHostKey()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "key")
	if err := os.WriteFile(path, []byte(host.private), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("ssh-keygen", "-y", "-f", path).Output()
	if err != nil {
		t.Fatalf("ssh-keygen cannot read the private key: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != host.public {
		t.Fatalf("public key derived by ssh-keygen %q, want %q", got, host.public)
	}
}

func TestScannedKeyMatches(t *testing.T) {
	expected := "ssh-ed25519 AAAAexpected"
	tests := []struct {
		name    string
		scanned string
		want    bool
	}{
		{"same key", "# 10.0.0.1:22 SSH-2.0-OpenSSH\n10.0.0.1 ssh-ed25519 AAAAexpected\n", true},
		{"other key", "10.0.0.1 ssh-ed25519 AAAAother\n", false},
		{"same blob, other type", "10.0.0.1 ssh-rsa AAAAexpected\n", false},
		{"only a comment", "# 10.0.0.1 ssh-ed25519 AAAAexpected\n", false},
		{"empty", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scannedKeyMatches([]byte(tc.scanned), expected); got != tc.want {
				t.Fatalf("scannedKeyMatches = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env")
	content := "# comment\nOVH_ENDPOINT=ovh-eu\nOVH_APPLICATION_KEY=\"quoted\"\nOVH_APPLICATION_SECRET='single'\n\nnot a pair\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	env, err := loadEnvFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"OVH_ENDPOINT": "ovh-eu", "OVH_APPLICATION_KEY": "quoted", "OVH_APPLICATION_SECRET": "single"}
	for k, v := range want {
		if env[k] != v {
			t.Errorf("%s = %q, want %q", k, env[k], v)
		}
	}
	if len(env) != len(want) {
		t.Errorf("parsed %d keys, want %d: %v", len(env), len(want), env)
	}

	missing, err := loadEnvFile(filepath.Join(t.TempDir(), "absent"))
	if err != nil || len(missing) != 0 {
		t.Fatalf("a missing file gave %v, %v; want an empty map", missing, err)
	}
}

func TestWriteEnvFileKeepsOtherLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env")
	if err := os.WriteFile(path, []byte("# keep me\nOVH_ENDPOINT=old\nEXTRA=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeEnvFile(path, map[string]string{"OVH_ENDPOINT": "ovh-eu", "OVH_CONSUMER_KEY": "ck"}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# keep me\nOVH_ENDPOINT=ovh-eu\nEXTRA=1\nOVH_CONSUMER_KEY=ck\n"
	if string(got) != want {
		t.Fatalf("file is %q, want %q", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o, want 600", info.Mode().Perm())
	}
}

func TestMatchImages(t *testing.T) {
	images := []Image{{ID: "1", Name: "Ubuntu 24.04"}, {ID: "2", Name: "Ubuntu 22.04"}, {ID: "3", Name: "Debian 12"}}
	if got := matchImages(images, "ubuntu 24"); len(got) != 1 || got[0].ID != "1" {
		t.Errorf("case-insensitive match gave %v", got)
	}
	if got := matchImages(images, "ubuntu"); len(got) != 2 {
		t.Errorf("ambiguous match gave %v, want both Ubuntu images", got)
	}
	if got := matchImages(images, "arch"); len(got) != 0 {
		t.Errorf("no match gave %v", got)
	}
}

func TestRenderScriptNeedsShebang(t *testing.T) {
	path := filepath.Join(t.TempDir(), "script.sh")
	if err := os.WriteFile(path, []byte("echo no shebang\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := renderScript(path, scriptVars{}); err == nil {
		t.Fatal("a script without a shebang was rendered")
	}
}
