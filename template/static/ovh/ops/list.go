package main

import (
	"flag"
	"fmt"
	"strings"
)

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	credentials := fs.String("credentials-file", "", "default: "+credentialsFile+" in the repo")
	_ = fs.Parse(args)

	client, err := newClient(*credentials)
	if err != nil {
		fatal(err)
	}

	var serviceNames []string
	if err := client.Get("/vps", &serviceNames); err != nil {
		fatal(err)
	}

	for _, name := range serviceNames {
		var addrs []string
		if err := client.Get(fmt.Sprintf("/vps/%s/ips", name), &addrs); err != nil {
			fatal(err)
		}
		fmt.Printf("%s\t%s\n", name, strings.Join(addrs, ", "))
	}
}
