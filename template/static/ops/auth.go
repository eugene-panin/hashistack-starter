package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/ovh/go-ovh/ovh"
)

var consumerKeyRules = []struct{ Method, Path string }{
	{"GET", "/vps"},
	{"GET", "/vps/*/ips"},
	{"GET", "/vps/*/images/available"},
	{"GET", "/vps/*/images/available/*"},
	{"GET", "/vps/*/tasks"},
	{"GET", "/vps/*/tasks/*"},
	{"POST", "/vps/*/rebuild"},
	{"POST", "/vps/*/getConsoleUrl"},
	{"POST", "/vps/*/reboot"},
	{"POST", "/vps/*/start"},
	{"POST", "/vps/*/stop"},
}

func cmdAuth(args []string) {
	fs := flag.NewFlagSet("auth", flag.ExitOnError)
	endpoint := fs.String("endpoint", "ovh-eu", "")
	appKey := fs.String("app-key", "", "from https://eu.api.ovh.com/createApp/ (or OVH_APPLICATION_KEY in the credentials file)")
	appSecret := fs.String("app-secret", "", "from https://eu.api.ovh.com/createApp/ (or OVH_APPLICATION_SECRET)")
	credentials := fs.String("credentials-file", "", "default: "+credentialsFile+" in the repo")
	_ = fs.Parse(args)

	path, err := repoPath(*credentials, credentialsFile)
	if err != nil {
		fatal(err)
	}
	env, err := loadEnvFile(path)
	if err != nil {
		fatal(err)
	}
	if *appKey == "" {
		*appKey = env["OVH_APPLICATION_KEY"]
	}
	if *appSecret == "" {
		*appSecret = env["OVH_APPLICATION_SECRET"]
	}
	if *appKey == "" || *appSecret == "" {
		fmt.Fprintf(os.Stderr, "missing --app-key/--app-secret (or OVH_APPLICATION_KEY/OVH_APPLICATION_SECRET in %s)\n"+
			"create an Application first at https://eu.api.ovh.com/createApp/ — that part can't be automated,\n"+
			"it's the root credential and OVH only issues it through a logged-in browser session\n", path)
		os.Exit(1)
	}

	client, err := ovh.NewClient(*endpoint, *appKey, *appSecret, "")
	if err != nil {
		fatal(err)
	}

	ck := client.NewCkRequest()
	for _, rule := range consumerKeyRules {
		ck.AddRule(rule.Method, rule.Path)
	}

	state, err := ck.Do()
	if err != nil {
		fatal(err)
	}

	fmt.Fprintln(os.Stderr, "open this URL, log in, and confirm:")
	fmt.Println(state.ValidationURL)
	fmt.Fprint(os.Stderr, "\npress Enter once confirmed to save the consumer key... ")
	if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
		fatal(fmt.Errorf("consumer key not saved, confirmation not read: %w", err))
	}

	if err := writeEnvFile(path, map[string]string{
		"OVH_ENDPOINT":           *endpoint,
		"OVH_APPLICATION_KEY":    *appKey,
		"OVH_APPLICATION_SECRET": *appSecret,
		"OVH_CONSUMER_KEY":       state.ConsumerKey,
	}); err != nil {
		fatal(err)
	}
	fmt.Fprintf(os.Stderr, "saved to %s\n", path)
}
