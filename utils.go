package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
)

func getEnv(key, defaultValue string) string {
	value, err := os.LookupEnv(key)
	if !err {
		return defaultValue
	}
	return value
}

func readFile(filePath string) []byte {
	f, ioErr := ioutil.ReadFile(filePath)
	if ioErr != nil {
		log.Fatal().
			Err(ioErr).
			Msgf("Failed to read a file. Expected path: %s.", filePath)
	}
	return f
}

func createAllowResponse(uid string) (response AdmissionResponse) {
	response = AdmissionResponse{
		APIVersion: "admission.k8s.io/v1",
		Kind:       "AdmissionReview",
		Response: Response{
			UID:     uid,
			Allowed: true,
		},
	}
	return
}

func createPatchResponse(uid string, patches []Patch) (response AdmissionResponse) {
	response = AdmissionResponse{
		APIVersion: "admission.k8s.io/v1",
		Kind:       "AdmissionReview",
		Response: Response{
			UID:     uid,
			Allowed: true,
		},
	}

	patchJson, err := json.Marshal(patches)
	if err != nil {
		log.Error().
			Err(err).
			Msgf("invalid patch object received. Cannot marshal theå patch object. Error: %s", err)
	}

	log.Debug().
		Msgf("Response to admission review: %v", patches)
	response.Response.Patch = base64.StdEncoding.EncodeToString(patchJson)
	response.Response.PatchType = "JSONPatch"

	return
}

func filterSts(references []OwnerReference) *OwnerReference {
	for i := range references {
		if references[i].Kind == "StatefulSet" {
			return &references[i]
		}
	}
	return nil
}
