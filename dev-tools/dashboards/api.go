package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var importAPI = "/api/kibana/import/dashboards"
var exportAPI = "/api/kibana/export/dashboards"

func makeURL(url, path string, params url.Values) string {

	if len(params) == 0 {
		return url + path
	}

	return strings.Join([]string{url, path, "?", params.Encode()}, "")
}

func ImportDashboards(client *http.Client, conn string, file string) (string, error) {

	// read json file
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Set("force", "true")            //overwrite the existing dashboards
	params.Add("exclude", "index-pattern") //don't import the index pattern

	fullURL := makeURL(conn, importAPI, params)

	// build the request
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(content))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-version", "6.0.0-alpha2")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	fmt.Println("Request: %s", req.URL)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return string(body), fmt.Errorf("HTTP POST %s fails with %s", fullURL, resp.Status)
	}
	return string(body), err

}

func ExportDashboards(client *http.Client, conn string, dashboards []string, out string) error {

	params := url.Values{}

	for _, dashboard := range dashboards {
		params.Add("dashboard", dashboard)
	}

	fullURL := makeURL(conn, exportAPI, params)
	fmt.Println(fullURL)

	req, err := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("kbn-version", "6.0.0-alpha2")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	fmt.Println(resp.Status)

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP GET %s fails with %s", fullURL, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fail to read response %s", err)
	}

	err = ioutil.WriteFile(out, body, 0666)

	fmt.Printf("Check %s file\n", out)
	return err
}

func main() {

	command := flag.String("cmd", "import", "import/export command")
	kibanaURL := flag.String("kibana", "http://localhost:5601", "Kibana URL")
	fileOutput := flag.String("file", "output.json", "File name")

	flag.Parse()

	args := flag.Args() //only for export

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}

	client := &http.Client{Transport: transCfg}

	if *command == "export" {
		err := ExportDashboards(client, *kibanaURL, args, *fileOutput)
		if err != nil {
			fmt.Printf("ERROR: fail to export the dashboards: %s\n", err)
		}
	} else if *command == "import" {
		res, err := ImportDashboards(client, *kibanaURL, *fileOutput)
		if err != nil {
			fmt.Printf("ERROR: fail to import the dashboards: %s\n", err)
			fmt.Println(res)
		}
	} else {
		fmt.Printf("Unknown command %s\n", *command)
	}
}
