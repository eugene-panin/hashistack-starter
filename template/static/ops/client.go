package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ovh/go-ovh/ovh"
)

const (
	configFile      = "ovh-stack.hcl"
	credentialsFile = "secrets/ovh-api.env"
)

var envKeyOrder = []string{
	"OVH_ENDPOINT",
	"OVH_APPLICATION_KEY",
	"OVH_APPLICATION_SECRET",
	"OVH_CONSUMER_KEY",
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, configFile)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%s not found here or in any parent directory; run ovhctl inside the repo", configFile)
		}
		dir = parent
	}
}

func repoPath(explicit, rel string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, rel), nil
}

func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}

func loadEnvFile(path string) (map[string]string, error) {
	env := map[string]string{}
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return env, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		env[strings.TrimSpace(key)] = unquote(strings.TrimSpace(value))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return env, nil
}

func writeEnvFile(path string, updates map[string]string) error {
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	written := map[string]bool{}
	var lines []string
	if len(raw) > 0 {
		for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
			key, _, found := strings.Cut(line, "=")
			key = strings.TrimSpace(key)
			if value, ok := updates[key]; found && ok && !strings.HasPrefix(key, "#") {
				line = key + "=" + value
				written[key] = true
			}
			lines = append(lines, line)
		}
	}
	for _, key := range envKeyOrder {
		if value, ok := updates[key]; ok && !written[key] {
			lines = append(lines, key+"="+value)
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") && path != "~" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return home + path[1:]
}

func promptIndex(count int) (int, error) {
	fmt.Print("select a number: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("read selection: %w", err)
	}
	line = strings.TrimSpace(line)
	idx, err := strconv.Atoi(line)
	if err != nil || idx < 1 || idx > count {
		return 0, fmt.Errorf("invalid selection %q", line)
	}
	return idx, nil
}

func newClient(explicitCredentials string) (*ovh.Client, error) {
	path, err := repoPath(explicitCredentials, credentialsFile)
	if err != nil {
		return nil, err
	}
	env, err := loadEnvFile(path)
	if err != nil {
		return nil, err
	}

	values := map[string]string{}
	var missing []string
	for _, key := range envKeyOrder {
		v := os.Getenv(key)
		if v == "" {
			v = env[key]
		}
		if v == "" {
			missing = append(missing, key)
		}
		values[key] = v
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing OVH credentials: %s (set in %s or the environment)",
			strings.Join(missing, ", "), path)
	}

	return ovh.NewClient(values["OVH_ENDPOINT"], values["OVH_APPLICATION_KEY"],
		values["OVH_APPLICATION_SECRET"], values["OVH_CONSUMER_KEY"])
}

func persistedServiceNamePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", filepath.Base(root), "service-name"), nil
}

func resolveServiceName(client *ovh.Client, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	if v := os.Getenv("OVHCTL_SERVICE_NAME"); v != "" {
		fmt.Fprintf(os.Stderr, "using OVHCTL_SERVICE_NAME: %s\n", v)
		return v, nil
	}

	if path, err := persistedServiceNamePath(); err == nil {
		if saved, err := os.ReadFile(path); err == nil {
			name := strings.TrimSpace(string(saved))
			if name != "" {
				fmt.Fprintf(os.Stderr, "using saved selection (%s): %s\n", path, name)
				return name, nil
			}
		}
	}

	var serviceNames []string
	if err := client.Get("/vps", &serviceNames); err != nil {
		return "", err
	}

	switch len(serviceNames) {
	case 0:
		return "", fmt.Errorf("no VPS found on this account")
	case 1:
		fmt.Fprintf(os.Stderr, "no --service-name given, using the only VPS on this account: %s\n", serviceNames[0])
		return serviceNames[0], nil
	default:
		return "", fmt.Errorf("multiple VPS on this account, run 'ovhctl use' or pass --service-name explicitly: %s",
			strings.Join(serviceNames, ", "))
	}
}
