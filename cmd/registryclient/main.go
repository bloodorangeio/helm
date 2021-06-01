package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	rx "helm.sh/helm/v3/pkg/registryx"
)

var (
	host      = os.Getenv("REGISTRY_HOST")
	namespace = os.Getenv("REGISTRY_NAMESPACE")
	user      = os.Getenv("REGISTRY_USERNAME")
	pass      = os.Getenv("REGISTRY_PASSWORD")
	chartPath = os.Getenv("CHART_PATH")
)

func usage() {
	fmt.Println("Must set the following env vars: " +
		"REGISTRY_HOST, REGISTRY_NAMESPACE, REGISTRY_USERNAME, " +
		"REGISTRY_PASSWORD, CHART_PATH")
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if host == "" || namespace == "" || user == "" || pass == "" {
		usage()
	}

	// Create client
	client, err := rx.NewClient()
	check(err)

	// Login to registry
	_, err = client.Login(host, rx.LoginOptBasicAuth(user, pass))
	check(err)

	// Load chart
	chartData, err := ioutil.ReadFile(chartPath)
	check(err)

	// Push chart
	result, err := client.Push(chartData, path.Join(host, namespace))
	check(err)

	fmt.Printf("Pushed %s\n", result.Ref)
	fmt.Printf("Pushed %s\n", result.RefWithDigest)
}
