package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

var CLIENT = &http.Client{}
var TOKEN string

func init() {
	// Load token
	token, tErr := ioutil.ReadFile(CONFIG.K8s.Token)
	if tErr != nil {
		log.Error().
			Err(tErr).
			Msgf("Cannot open a token file: %s", CONFIG.K8s.Token)
	}
	TOKEN = string(token)

	// Load ca cert file
	caCert, cErr := ioutil.ReadFile(CONFIG.K8s.Tls.CaCert)
	if cErr != nil {
		log.Fatal().
			Err(cErr).
			Msgf("Cannot open a CA cert file: %s", CONFIG.K8s.Tls.CaCert)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	CLIENT = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
}

func filterSts(references []OwnerReference) *OwnerReference {
	for i := range references {
		if references[i].Kind == "StatefulSet" {
			return &references[i]
		}
	}
	return nil
}

func getSts(namespace string, name string) *Statefulset {

	// get api path
	url := fmt.Sprintf(
		"%s/apis/apps/v1/namespaces/%s/statefulsets/%s", CONFIG.K8s.URL, namespace, name)
	req, _ := http.NewRequest("GET", url, nil)

	// add authorization header
	req.Header.Add("Authorization", `Bearer `+TOKEN)
	resp, connectionErr := CLIENT.Do(req)

	if connectionErr != nil {
		log.Error().
			Err(connectionErr).
			Msgf("Request to %s failed. Cause %s", url, connectionErr.Error())
		return nil
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Error().
			Msgf("Request to %s failed. Returned with status %s", url, resp.Status)
		return nil
	}

	// parse the response
	sts := &Statefulset{}
	decodeErr := json.NewDecoder(resp.Body).Decode(sts)
	if decodeErr != nil {
		log.Error().
			Err(decodeErr).
			Msgf("Failed to decode Statefulset %s by struct", resp.Body)
		return nil
	}
	return sts
}
