package main

import (
	"github.com/go-playground/assert/v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
)

func TestGetSts(t *testing.T) {
	// TODO: Use K8s Mock Master API for testing. Currently, this expects sts.yaml in test dir is running in cluster
	expect := readExpectedSts()
	actual := getSts(expect.Metadata.Namespace, expect.Metadata.Name)
	assert.NotEqual(t, nil, actual)
	assert.Equal(t, expect.Spec.Replicas, actual.Spec.Replicas)
}

func TestGetStsNotFound(t *testing.T) {
	// TODO: Use K8s Mock Master API for testing. Currently, this expects sts.yaml in test dir is running in cluster
	actual := getSts("INVALID-NAMESPACE", "INVALID-NAME")
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
