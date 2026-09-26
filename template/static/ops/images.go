package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ovh/go-ovh/ovh"
)

type Image struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func cmdImages(args []string) {
	fs := flag.NewFlagSet("images", flag.ExitOnError)
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

	images, err := listImages(client, name)
	if err != nil {
		fatal(err)
	}
	for _, img := range images {
		fmt.Printf("%s\t%s\n", img.ID, img.Name)
	}
}

func listImages(client *ovh.Client, serviceName string) ([]Image, error) {
	var ids []string
	if err := client.Get(fmt.Sprintf("/vps/%s/images/available", serviceName), &ids); err != nil {
		return nil, err
	}

	images := make([]Image, 0, len(ids))
	for _, id := range ids {
		var img Image
		path := fmt.Sprintf("/vps/%s/images/available/%s", serviceName, id)
		if err := client.Get(path, &img); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func resolveImageID(client *ovh.Client, serviceName, explicitID, nameSubstring string) (string, error) {
	if explicitID != "" {
		return explicitID, nil
	}

	images, err := listImages(client, serviceName)
	if err != nil {
		return "", err
	}

	if nameSubstring == "" {
		for i, img := range images {
			fmt.Printf("%d) %s\t%s\n", i+1, img.ID, img.Name)
		}
		idx, err := promptIndex(len(images))
		if err != nil {
			return "", err
		}
		return images[idx-1].ID, nil
	}

	matches := matchImages(images, nameSubstring)
	switch len(matches) {
	case 1:
		fmt.Fprintf(os.Stderr, "matched image: %s (%s)\n", matches[0].Name, matches[0].ID)
		return matches[0].ID, nil
	case 0:
		var available []string
		for _, img := range images {
			available = append(available, img.Name)
		}
		return "", fmt.Errorf("no image name matches %q, available: %s", nameSubstring, strings.Join(available, ", "))
	default:
		var names []string
		for _, img := range matches {
			names = append(names, fmt.Sprintf("%s (%s)", img.Name, img.ID))
		}
		return "", fmt.Errorf("%q matches multiple images, be more specific: %s", nameSubstring, strings.Join(names, ", "))
	}
}

func matchImages(images []Image, nameSubstring string) []Image {
	var matches []Image
	needle := strings.ToLower(nameSubstring)
	for _, img := range images {
		if strings.Contains(strings.ToLower(img.Name), needle) {
			matches = append(matches, img)
		}
	}
	return matches
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
