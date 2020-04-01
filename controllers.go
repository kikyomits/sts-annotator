package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
)

func health(c *gin.Context) {
	response := Health{
		Health: true,
	}
	c.JSON(http.StatusOK, response)
}

func annotateStsPod(c *gin.Context) {
	// bind a request body to struct
	var payload AdmissionRequest
	err := c.BindJSON(&payload)
	if err != nil {
		log.Error().
			Err(err).
			Msgf("Internal Server Error. Failed to bind AdmissionRequest to requested payload")
		c.AbortWithStatusJSON(http.StatusInternalServerError, createAllowResponse("invalid-request"))
	}

	if payload.Request.UID == "" || payload.Request.Namespace == "" {
		log.Error().
			Msgf("Failed to parse request. Received invalid request body %v", payload)
		c.AbortWithStatusJSON(http.StatusBadRequest, createAllowResponse("invalid-request"))
		return
	}

	uid, namespace, obj := payload.Request.UID, payload.Request.Namespace, payload.Request.Object
	podName := payload.Request.Object.Metadata.Name
	logMetadata := fmt.Sprintf(
		"uid: %s namespace: %s pod: %s", uid, namespace, podName)

	// if this POD Create request is not for Statefulset, just allow
	stsOwnerRef := filterSts(obj.Metadata.OwnerReferences)
	if stsOwnerRef == nil {
		log.Debug().Msgf("Skip. Statefulset is not found in OwnerReferences %s", logMetadata)
		c.AbortWithStatusJSON(http.StatusOK, createAllowResponse(uid))
		return
	}

	// if this POD Create request is for Statefulset, add annotations
	log.Info().Msgf("Add annotation to Statefulset Pod. %s", logMetadata)
	podNameSplit := strings.Split(podName, "-")
	podIndex := podNameSplit[len(podNameSplit)-1]
	sts := getSts(namespace, stsOwnerRef.Name)

	// TODO: more precise error information woold be required.
	if sts == nil {
		log.Error().
			Msgf("Statefulset %s not found. %s", stsOwnerRef.Name, logMetadata)
		c.AbortWithStatusJSON(http.StatusInternalServerError, createAllowResponse(uid))
		return
	}

	var patches []Patch
	constant := newConstant()
	// if the pod doesn't have any annotation field, must create the field `annotations` first.
	if obj.Metadata.Annotations == nil {
		empty := make(map[string]string)
		patches = append(patches, createPatch(constant.AnnotationsPath, empty))
	}

	podReplicas := strconv.Itoa(sts.Spec.Replicas)
	patches = append(patches, createPatch(constant.PodReplicasPath, podReplicas))
	patches = append(patches, createPatch(constant.PodIndexPath, podIndex))

	c.AbortWithStatusJSON(http.StatusOK, createPatchResponse(uid, patches))
	return
}

func createPatch(path string, value interface{}) (patch Patch) {
	patch = newPatch()
	patch.Path = path
	patch.Value = value
	return
}
