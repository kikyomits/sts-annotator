package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
)

type K8sService struct {
	baseUrl string
	token   string
	client  http.Client
}

func newK8sService(config *Config) K8sService {
	token := string(readFile(config.K8s.Token))

	// Load ca cert file
	caCert := readFile(config.K8s.Tls.CaCert)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create http client
	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	return K8sService{baseUrl: config.K8s.URL, token: token, client: httpClient}
}

func filterSts(references []OwnerReference) *OwnerReference {
	for i := range references {
		if references[i].Kind == "StatefulSet" {
			return &references[i]
		}
	}
	return nil
}

func (k8s K8sService) getSts(namespace string, name string) *Statefulset {
	// get api path
	url := fmt.Sprintf(
		"%s/apis/apps/v1/namespaces/%s/statefulsets/%s", k8s.baseUrl, namespace, name)
	req, _ := http.NewRequest("GET", url, nil)

	// add authorization header
	authorization := fmt.Sprintf("Bearer %s", k8s.token)
	req.Header.Add("Authorization", authorization)
	resp, connectionErr := k8s.client.Do(req)

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
