package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/rs/zerolog/log"
)

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
