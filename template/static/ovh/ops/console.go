package main

import (
	"flag"
	"fmt"
)

func cmdConsole(args []string) {
	fs := flag.NewFlagSet("console", flag.ExitOnError)
	serviceName := fs.String("service-name", "", "defaults to the account's only VPS, if there's exactly one")
	credentials := fs.String("credentials-file", "", "default: "+credentialsFile+" in the repo")
	_ = fs.Parse(args)

	client, err := newClient(*credentials)
	if err != nil {
		fatal(err)
	}
	name, err := resolveServiceName(client, *serviceName)
	if err != nil {
		fatal(err)
	}

	var url string
	path := fmt.Sprintf("/vps/%s/getConsoleUrl", name)
	if err := client.Post(path, nil, &url); err != nil {
		fatal(err)
	}
	fmt.Println(url)
}
