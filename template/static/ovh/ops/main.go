package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "auth":
		cmdAuth(os.Args[2:])
	case "list":
		cmdList(os.Args[2:])
	case "use":
		cmdUse(os.Args[2:])
	case "images":
		cmdImages(os.Args[2:])
	case "rebuild":
		cmdRebuild(os.Args[2:])
	case "host":
		cmdHost(os.Args[2:])
	case "console":
		cmdConsole(os.Args[2:])
	case "tasks":
		cmdTasks(os.Args[2:])
	case "reboot", "start", "stop":
		cmdPower(os.Args[1], os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  ovhctl auth    --app-key <key> --app-secret <secret> # request+save a scoped consumer key
  ovhctl list                                         # list VPS service names on this account
  ovhctl use     [--service-name <name>]              # persist a default (~/.config/<repo>/service-name), like kubectl's current-context
  ovhctl host    [--service-name <name>]              # print the IPv4 address, for feeding into the Ansible inventory
  ovhctl images  --service-name <name>
  ovhctl rebuild --service-name <name> --image-id <id> --public-key <path> [--script <path>] [--confirm] [--trust-scanned-host-key]
  ovhctl console --service-name <name>                # break-glass KVM/VNC URL, works even if SSH is locked out
  ovhctl tasks   --service-name <name>                # read-only: what OVH actually has on record for this VPS
  ovhctl reboot  --service-name <name>
  ovhctl start   --service-name <name>
  ovhctl stop    --service-name <name>`)
}
