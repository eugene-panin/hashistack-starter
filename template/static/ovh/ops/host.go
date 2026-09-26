package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func cmdHost(args []string) {
	fs := flag.NewFlagSet("host", flag.ExitOnError)
	serviceName := fs.String("service-name", "", "defaults to the saved selection")
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

	var addrs []string
	if err := client.Get(fmt.Sprintf("/vps/%s/ips", name), &addrs); err != nil {
		fatal(err)
	}

	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip != nil && ip.To4() != nil {
			fmt.Println(addr)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "%s has no IPv4 address\n", name)
	os.Exit(1)
}
