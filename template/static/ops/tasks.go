package main

import (
	"flag"
	"fmt"
)

func cmdTasks(args []string) {
	fs := flag.NewFlagSet("tasks", flag.ExitOnError)
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

	var taskIDs []int64
	if err := client.Get(fmt.Sprintf("/vps/%s/tasks", name), &taskIDs); err != nil {
		fatal(err)
	}

	for _, id := range taskIDs {
		var t Task
		if err := client.Get(fmt.Sprintf("/vps/%s/tasks/%d", name, id), &t); err != nil {
			fatal(err)
		}
		fmt.Printf("%d\t%s\t%s\t%s\n", t.ID, t.Type, t.State, t.Date)
	}
}
