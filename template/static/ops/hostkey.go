package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"strings"

	"golang.org/x/crypto/ssh"
)

type hostKey struct {
	private string
	public  string
}

func newHostKey() (hostKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return hostKey{}, err
	}
	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return hostKey{}, err
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return hostKey{}, err
	}
	return hostKey{
		private: string(pem.EncodeToMemory(block)),
		public:  strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub))),
	}, nil
}

func scannedKeyMatches(scanned []byte, expected string) bool {
	want := strings.Fields(expected)
	if len(want) < 2 {
		return false
	}
	for _, line := range bytes.Split(scanned, []byte("\n")) {
		fields := strings.Fields(string(line))
		if len(fields) >= 3 && !strings.HasPrefix(fields[0], "#") && fields[1] == want[0] && fields[2] == want[1] {
			return true
		}
	}
	return false
}
