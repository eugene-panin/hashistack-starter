package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cmdUse(args []string) {
	fs := flag.NewFlagSet("use", flag.ExitOnError)
	explicit := fs.String("service-name", "", "save this name directly, skip the prompt")
	credentials := fs.String("credentials-file", "", "default: "+credentialsFile+" in the repo")
	_ = fs.Parse(args)

	if *explicit != "" {
		saveSelection(*explicit)
		return
	}

	client, err := newClient(*credentials)
	if err != nil {
		fatal(err)
	}

	var serviceNames []string
	if err := client.Get("/vps", &serviceNames); err != nil {
		fatal(err)
	}

	switch len(serviceNames) {
	case 0:
		fatal(fmt.Errorf("no VPS found on this account"))
	case 1:
		saveSelection(serviceNames[0])
		return
	}

	for i, name := range serviceNames {
		var addrs []string
		if err := client.Get(fmt.Sprintf("/vps/%s/ips", name), &addrs); err != nil {
			fatal(err)
		}
		fmt.Printf("%d) %s\t%s\n", i+1, name, strings.Join(addrs, ", "))
	}

	idx, err := promptIndex(len(serviceNames))
	if err != nil {
		fatal(err)
	}

	saveSelection(serviceNames[idx-1])
}

func saveSelection(name string) {
	path, err := persistedServiceNamePath()
	if err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(path, []byte(name+"\n"), 0o600); err != nil {
		fatal(err)
	}
	fmt.Fprintf(os.Stderr, "saved %s to %s — every other command will default to it\n", name, path)
}
