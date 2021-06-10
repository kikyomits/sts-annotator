package main

import (
	"encoding/json"
	"github.com/go-playground/assert/v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var NotFound = "NOT_FOUND"
var BadRequest = "BadRequest"

func TestGetSts(t *testing.T) {
	expect := readExpectedSts()
	bytes, _ := json.Marshal(expect)
	ts := buildTS(&bytes)
	defer ts.Close()

	svc := buildMockSvc(ts.URL)
	actual := svc.getSts(expect.Metadata.Namespace, expect.Metadata.Name)
	assert.NotEqual(t, nil, actual)
	assert.Equal(t, expect.Spec.Replicas, actual.Spec.Replicas)
}

func TestGetStsNotFound(t *testing.T) {
	expect := readExpectedSts()
	bytes, _ := json.Marshal(expect)
	ts := buildTS(&bytes)
	defer ts.Close()

	svc := buildMockSvc(ts.URL)
	actual := svc.getSts(expect.Metadata.Namespace, NotFound)
	assert.Equal(t, nil, actual)
}

func TestFilterSts(t *testing.T) {
	ownerNonSts := OwnerReference{Kind: "Deployment"}
	resultNonSts := filterSts(append([]OwnerReference{}, ownerNonSts))
	assert.Equal(t, nil, resultNonSts)

	ownerSts := OwnerReference{Kind: "StatefulSet"}
	resultSts := filterSts(append([]OwnerReference{}, ownerSts))
	assert.Equal(t, ownerSts, resultSts)
}

func readExpectedSts() (expect Statefulset) {
	expect = Statefulset{}
	yamlFile, _ := ioutil.ReadFile("test/sts.yaml")
	_ = yaml.Unmarshal(yamlFile, &expect)
	return
}

func buildTS(body *[]byte) (ts *httptest.Server) {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), NotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write(*body)
		} else if strings.Contains(r.URL.String(), BadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(*body)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(*body)
		}
	}))
	return
}

func buildMockSvc(url string) (svc K8sService) {
	config := initTestConfig(url)
	svc = newK8sService(config)
	return
}
