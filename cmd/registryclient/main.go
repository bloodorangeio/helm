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
	fmt.Printf("Attempting to login to %s ...\n", host)
	_, err = client.Login(host, rx.LoginOptBasicAuth(user, pass))
	check(err)

	// Build a valid Helm OCI reference
	ref, err := rx.BuildRefFromChartData(path.Join(host, namespace), chartData)
	check(err)

	// Build options for push operation
	var pushOpts []rx.PushOption
	if hasProv {
		pushOpts = append(pushOpts, rx.PushOptProvData(provData))
	}

	// Push chart (and prov if specified)
	fmt.Printf("Attempting push to %s ...\n", ref)
	pushResult, err := client.Push(chartData, ref, pushOpts...)
	check(err)
	chartName := pushResult.Chart.Meta.Name
	chartVersion := pushResult.Chart.Meta.Version
	fmt.Printf("Manifest digest:    %s\n", pushResult.Manifest.Digest)
	fmt.Printf("Manifest size:      %d\n", pushResult.Manifest.Size)
	fmt.Printf("Config digest:      %s\n", pushResult.Config.Digest)
	fmt.Printf("Config size:        %d\n", pushResult.Config.Size)
	fmt.Printf("Chart layer digest: %s\n", pushResult.Chart.Digest)
	fmt.Printf("Chart layer size:   %d\n", pushResult.Chart.Size)
	fmt.Printf("Chart name:         %s\n", chartName)
	fmt.Printf("Chart version:      %s\n", chartVersion)
	if hasProv {
		fmt.Printf("Prov layer digest:  %s\n", pushResult.Prov.Digest)
		fmt.Printf("Prov layer size:    %d\n", pushResult.Prov.Size)
	}

	// Build options for pull operation
	var pullOpts []rx.PullOption
	if hasProv {
		pullOpts = append(pullOpts, rx.PullOptWithProv(true))
	}

	// Pull chart (and prov if specified)
	fmt.Printf("Attempting pull from %s ...\n", ref)
	pullResult, err := client.Pull(ref, pullOpts...)
	check(err)
	chartName = pullResult.Chart.Meta.Name
	chartVersion = pullResult.Chart.Meta.Version
	fmt.Printf("Manifest digest:    %s\n", pullResult.Manifest.Digest)
	fmt.Printf("Manifest size:      %d\n", pullResult.Manifest.Size)
	fmt.Printf("Config digest:      %s\n", pullResult.Config.Digest)
	fmt.Printf("Config size:        %d\n", pullResult.Config.Size)
	fmt.Printf("Chart layer digest: %s\n", pullResult.Chart.Digest)
	fmt.Printf("Chart layer size:   %d\n", pullResult.Chart.Size)
	fmt.Printf("Chart name:         %s\n", chartName)
	fmt.Printf("Chart version:      %s\n", chartVersion)
	if hasProv {
		fmt.Printf("Prov layer digest:  %s\n", pullResult.Prov.Digest)
		fmt.Printf("Prov layer size:    %d\n", pullResult.Prov.Size)
	}

	// Create a directory to output download to
	os.MkdirAll("./output/", os.ModePerm)

	// Write data to files
	chartOutputPath := fmt.Sprintf("./output/%s-%s.tgz", chartName, chartVersion)
	fmt.Printf("Saving chart to %s ...\n", chartOutputPath)
	err = ioutil.WriteFile(chartOutputPath, pullResult.Chart.Data, 0644)
	check(err)
	if hasProv {
		provOutputPath := fmt.Sprintf("%s.prov", chartOutputPath)
		fmt.Printf("Saving prov to %s ...\n", provOutputPath)
		err = ioutil.WriteFile(provOutputPath, pullResult.Prov.Data, 0644)
		check(err)
	}

	// Logout from registry
	fmt.Printf("Attempting to logout from %s ...\n", host)
	_, err = client.Logout(host)
	check(err)
}
