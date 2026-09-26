package main

import (
	"flag"
	"fmt"
)

func cmdPower(action string, args []string) {
	fs := flag.NewFlagSet(action, flag.ExitOnError)
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

	var task Task
	path := fmt.Sprintf("/vps/%s/%s", name, action)
	if err := client.Post(path, nil, &task); err != nil {
		fatal(err)
	}
	fmt.Printf("task %d: %s\n", task.ID, task.State)
}
