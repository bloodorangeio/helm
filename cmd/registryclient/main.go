package main

import (
	"fmt"
	"io/ioutil"
	"os"

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

	// Load raw chart .tgz
	chartData, err := ioutil.ReadFile(chartPath)
	check(err)

	// Load associated .prov file if present
	var provData []byte
	var hasProv bool
	provPath := fmt.Sprintf("%s.prov", chartPath)
	if _, err := os.Stat(provPath); err == nil {
		provData, err = ioutil.ReadFile(provPath)
		check(err)
		hasProv = true
	}

	// Create client
	client, err := rx.NewClient()
	check(err)

	// Login to registry
	_, err = client.Login(host, rx.LoginOptBasicAuth(user, pass))
	check(err)

	// Build a valid Helm OCI reference
	ref, err := rx.BuildRefFromChartData(host, namespace, chartData)
	check(err)

	// Build options for push operation
	var pushOpts []rx.PushOption
	if hasProv {
		pushOpts = append(pushOpts, rx.PushOptProvData(provData))
	}

	// Push chart (and prov if specified)
	fmt.Printf("Attempting push to %s ...\n", ref)
	result, err := client.Push(chartData, ref, pushOpts...)
	check(err)
	fmt.Printf("Manifest digest:    %s\n", result.Manifest.Digest)
	fmt.Printf("Manifest size:      %d\n", result.Manifest.Size)
	fmt.Printf("Config digest:      %s\n", result.Config.Digest)
	fmt.Printf("Config size:        %d\n", result.Config.Size)
	fmt.Printf("Chart layer digest: %s\n", result.Chart.Digest)
	fmt.Printf("Chart layer size:   %d\n", result.Chart.Size)
	if hasProv {
		fmt.Printf("Prov layer digest:  %s\n", result.Prov.Digest)
		fmt.Printf("Prov layer size:    %d\n", result.Prov.Size)
	}
}
