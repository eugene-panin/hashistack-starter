package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ovh/go-ovh/ovh"
)

const (
	scriptFile        = "ops/templates/rebuild.sh"
	maxPollFailures   = 6
	keyscanAttempts   = 10
	keyscanRetryDelay = 3 * time.Second
)

type scriptVars struct {
	OpsUser    string
	PublicKey  string
	HostKey    string
	HostKeyPub string
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func (v scriptVars) preamble() string {
	var b strings.Builder
	fmt.Fprintf(&b, "OPS_USER=%s\n", shellQuote(v.OpsUser))
	fmt.Fprintf(&b, "PUBLIC_KEY=%s\n", shellQuote(v.PublicKey))
	fmt.Fprintf(&b, "HOST_KEY=%s\n", shellQuote(v.HostKey))
	fmt.Fprintf(&b, "HOST_KEY_PUB=%s\n", shellQuote(v.HostKeyPub))
	return b.String()
}

func renderScript(explicit string, vars scriptVars) (string, error) {
	path, err := repoPath(explicit, scriptFile)
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	body := string(raw)
	if !strings.HasPrefix(body, "#!") || !strings.Contains(body, "\n") {
		return "", fmt.Errorf("%s: expected a shebang on the first line", path)
	}
	shebangEnd := strings.Index(body, "\n") + 1
	return body[:shebangEnd] + vars.preamble() + body[shebangEnd:], nil
}

type RebuildRequest struct {
	DoNotSendPassword bool   `json:"doNotSendPassword"`
	ImageID           string `json:"imageId"`
	PostInstallScript string `json:"postInstallScript,omitempty"`
	PublicSSHKey      string `json:"publicSshKey,omitempty"`
}

type Task struct {
	ID    int64  `json:"id"`
	State string `json:"state"`
	Type  string `json:"type"`
	Date  string `json:"date"`
}

func cmdRebuild(args []string) {
	fs := flag.NewFlagSet("rebuild", flag.ExitOnError)
	serviceName := fs.String("service-name", "", "defaults to the account's only VPS, if there's exactly one")
	imageID := fs.String("image-id", "", "from 'ovhctl images'")
	imageName := fs.String("image-name", "", "substring match against 'ovhctl images' names (default: ovh-stack.hcl vps.image_name)")
	publicKeyPath := fs.String("public-key", "", "path to an SSH public key file (default: ovh-stack.hcl bootstrap.public_key_path)")
	scriptPath := fs.String("script", "", "bash script path (default: "+scriptFile+" in the repo)")
	configPath := fs.String("config", "", "project config (default: "+configFile+" in the repo)")
	credentials := fs.String("credentials-file", "", "default: "+credentialsFile+" in the repo")
	pollSeconds := fs.Int("poll-seconds", 10, "seconds between task status checks, at least 1")
	timeout := fs.Duration("timeout", time.Hour, "give up waiting for the rebuild task after this long")
	confirm := fs.Bool("confirm", false, "actually send the rebuild request; otherwise dry-run only")
	showScript := fs.Bool("show-script", false, "print the full rendered postInstallScript on dry-run, instead of just a summary")
	trustScanned := fs.Bool("trust-scanned-host-key", false, "if the host key after the rebuild is not the one ovhctl installed, trust it anyway")
	_ = fs.Parse(args)

	if *pollSeconds < 1 {
		fatal(fmt.Errorf("--poll-seconds must be at least 1, got %d", *pollSeconds))
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fatal(err)
	}
	if *publicKeyPath == "" {
		*publicKeyPath = cfg.Bootstrap.PublicKeyPath
	}
	if *imageID == "" && *imageName == "" && cfg.Vps != nil {
		*imageName = cfg.Vps.ImageName
	}

	client, err := newClient(*credentials)
	if err != nil {
		fatal(err)
	}
	name, err := resolveServiceName(client, *serviceName)
	if err != nil {
		fatal(err)
	}
	id, err := resolveImageID(client, name, *imageID, *imageName)
	if err != nil {
		fatal(err)
	}

	publicKeyBytes, err := os.ReadFile(expandHome(*publicKeyPath))
	if err != nil {
		fatal(err)
	}
	publicKey := strings.TrimSpace(string(publicKeyBytes))

	host, err := newHostKey()
	if err != nil {
		fatal(err)
	}

	req := RebuildRequest{
		DoNotSendPassword: true,
		ImageID:           id,
		PublicSSHKey:      publicKey,
	}
	script, err := renderScript(*scriptPath, scriptVars{
		OpsUser:    cfg.Bootstrap.OpsUser,
		PublicKey:  publicKey,
		HostKey:    host.private,
		HostKeyPub: host.public,
	})
	if err != nil {
		fatal(err)
	}
	req.PostInstallScript = script

	scOutput, scInstalled, scOK := checkScript(script)
	switch {
	case !scInstalled:
		fmt.Fprintln(os.Stderr, "shellcheck not installed, postInstallScript not checked")
	case !scOK:
		fmt.Fprintln(os.Stderr, "shellcheck found issues in the rendered script:")
		fmt.Fprintln(os.Stderr, scOutput)
		if *confirm {
			fatal(fmt.Errorf("refusing to send a script shellcheck flagged; fix it, or rerun without --confirm to inspect --show-script"))
		}
	}

	if !*confirm {
		fmt.Fprintln(os.Stderr, "DRY RUN — would POST the following and stop here:")
		fmt.Printf("POST /vps/%s/rebuild\n", name)
		fmt.Printf("  doNotSendPassword: %v\n", req.DoNotSendPassword)
		fmt.Printf("  imageId:           %s\n", req.ImageID)
		fmt.Printf("  publicSshKey:      %s\n", req.PublicSSHKey)
		fmt.Printf("  host key:          %s (a new one is generated on every run)\n", host.public)
		if *showScript {
			fmt.Println("  postInstallScript:")
			fmt.Println("--- rebuild.sh, rendered ---")
			fmt.Println(req.PostInstallScript)
			fmt.Println("--- end ---")
		} else {
			lines := strings.Count(req.PostInstallScript, "\n")
			sum := sha256.Sum256([]byte(req.PostInstallScript))
			fmt.Printf("  postInstallScript: %d lines, sha256 %x... (--show-script for the full text)\n", lines, sum[:4])
		}
		return
	}

	var task Task
	if err := client.Post(fmt.Sprintf("/vps/%s/rebuild", name), req, &task); err != nil {
		fatal(err)
	}
	fmt.Fprintf(os.Stderr, "rebuild started, task %d\n", task.ID)

	state, err := waitForTask(client, name, task.ID, time.Duration(*pollSeconds)*time.Second, *timeout)
	if err != nil {
		fatal(err)
	}
	if state != "done" {
		fatal(fmt.Errorf("rebuild task %d ended in state %s", task.ID, state))
	}
	if err := trustHostKey(client, name, host.public, *trustScanned); err != nil {
		fatal(err)
	}
}

func waitForTask(client *ovh.Client, serviceName string, id int64, poll, timeout time.Duration) (string, error) {
	path := fmt.Sprintf("/vps/%s/tasks/%d", serviceName, id)
	deadline := time.Now().Add(timeout)
	failures := 0
	for {
		var t Task
		if err := client.Get(path, &t); err != nil {
			failures++
			if failures >= maxPollFailures {
				return "", fmt.Errorf("poll task %d: %w", id, err)
			}
			fmt.Fprintf(os.Stderr, "poll task %d failed (%d/%d), retrying: %v\n", id, failures, maxPollFailures, err)
		} else {
			failures = 0
			fmt.Fprintf(os.Stderr, "task %d: %s\n", t.ID, t.State)
			switch t.State {
			case "done", "error", "cancelled":
				return t.State, nil
			}
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("task %d still running after %s; see 'ovhctl tasks'", id, timeout)
		}
		time.Sleep(poll)
	}
}

func trustHostKey(client *ovh.Client, serviceName, expected string, trustScanned bool) error {
	var addrs []string
	if err := client.Get(fmt.Sprintf("/vps/%s/ips", serviceName), &addrs); err != nil {
		return fmt.Errorf("list IPs to trust the host key: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir to update known_hosts: %w", err)
	}
	knownHosts := filepath.Join(home, ".ssh", "known_hosts")

	var failed []error
	for _, addr := range addrs {
		_ = exec.Command("ssh-keygen", "-R", addr).Run()

		scanned := keyscan(addr)
		if scanned == nil {
			failed = append(failed, fmt.Errorf("sshd on %s never answered ssh-keyscan; known_hosts not updated", addr))
			continue
		}

		entry := []byte(addr + " " + expected + "\n")
		if !scannedKeyMatches(scanned, expected) {
			if !trustScanned {
				failed = append(failed, fmt.Errorf("%s answers with a host key other than the one ovhctl installed; known_hosts not updated.\n"+
					"expected: %s\nscanned:\n%s\nrerun with --trust-scanned-host-key only if you know why they differ", addr, expected, scanned))
				continue
			}
			entry = scanned
		}

		if err := appendFile(knownHosts, entry); err != nil {
			failed = append(failed, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "trusted the host key of %s\n", addr)
	}
	return errors.Join(failed...)
}

func keyscan(addr string) []byte {
	for range keyscanAttempts {
		out, err := exec.Command("ssh-keyscan", "-T", "5", "-t", "ed25519", addr).Output()
		if err == nil && len(bytes.TrimSpace(out)) > 0 {
			return out
		}
		time.Sleep(keyscanRetryDelay)
	}
	return nil
}

func appendFile(path string, data []byte) (err error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	_, err = f.Write(data)
	return err
}
