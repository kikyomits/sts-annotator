package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var c = newConstant()
var url = c.V1 + c.Annotator
var method = "POST"

func setupTestRouter() (router *gin.Engine) {
	// RUn Gin server for unit test

	k8sRespBody, _ := json.Marshal(readExpectedSts())
	ts := buildTS(&k8sRespBody)
	config := initTestConfig(ts.URL)
	router = setupRouter(config)
	return
}

func TestHealth(t *testing.T) {
	// Verify health check endpoint
	// Expected: return 200.

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", c.V1+c.Healthz, nil)
	router := setupTestRouter()
	router.ServeHTTP(w, req)

	body, _ := ioutil.ReadAll(w.Body)
	expect := Health{Health: true}
	actual := Health{}
	_ = json.Unmarshal(body, &actual)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expect, actual)
}

func TestAnnotateStsPod(t *testing.T) {
	// Most standard case. Receive Statefulset Pod creation event.
	// This case expect the created pod doesn't have any annotations.
	// Expected: Return 200 with annotations.

	f := readFile("test/admissionRequest.json")
	w := sendRequest(f)

	actual := unmarshalActual(w)
	expect := unmarshalExpected(f)
	assertCommon(t, w.Code, expect, actual)
	assertPatches(t, expect, actual)
}

func TestAnnotateStsPodWithAnnotations(t *testing.T) {
	// Most standard case. Receive Statefulset Pod creation event.
	// This case expect the created pod has some annotations.
	// Expected: Return 200 with annotations.

	f := readFile("test/admissionRequestWithAnnotations.json")
	w := sendRequest(f)
	actual := unmarshalActual(w)
	expect := unmarshalExpected(f)
	assertCommon(t, w.Code, expect, actual)
	assertPatches(t, expect, actual)
}

func TestAnnotateNoStsPod(t *testing.T) {
	// Pod Creation Event (Other than Statefulset)
	// Expected Behavior: Skip adding annotation as the pod doesn't belong to Statefulset. Just return 200.
	f := readFile("test/admissionRequestNoSts.json")
	w := sendRequest(f)
	actual := unmarshalActual(w)
	expect := unmarshalExpected(f)
	assertCommon(t, w.Code, expect, actual)
	assert.Equal(t, "", actual.Response.Patch)
	assert.Equal(t, "", actual.Response.PatchType)
}

func TestAnnotateStsPodInvalidReq(t *testing.T) {
	// Pod Creation Event with invalid json.
	// Edge test case, just in case k8s's Admission Json object got changed.
	// Expected: Return 400.
	f := readFile("test/admissionRequestInvalid.json")
	w := sendRequest(f)
	actual := unmarshalActual(w)
	assertFailed(t, http.StatusBadRequest, w.Code, actual)
}

func TestAnnotateStsNotFound(t *testing.T) {
	// Pod Creation Event but the Statefulset has been gone somehow.
	// Edge test case.
	// Expected: Return 500

	f := readFile("test/admissionRequestStsNotFound.json")
	w := sendRequest(f)
	actual := unmarshalActual(w)
	assertFailed(t, http.StatusInternalServerError, w.Code, actual)
}

func assertCommon(t *testing.T, statusCode int, expect AdmissionRequest, actual AdmissionResponse) {
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, expect.Request.UID, actual.Response.UID)
	assert.Equal(t, true, actual.Response.Allowed)
}

func assertFailed(t *testing.T, expectCode int, actualCode int, actual AdmissionResponse) {
	assert.Equal(t, expectCode, actualCode)
	assert.Equal(t, true, actual.Response.Allowed)
	assert.Equal(t, "", actual.Response.Patch)
	assert.Equal(t, "", actual.Response.PatchType)
}

func assertPatches(
	t *testing.T,
	expect AdmissionRequest,
	actual AdmissionResponse) {
	assert.Equal(t, "jsonpatch", strings.ToLower(actual.Response.PatchType))

	patches := unmarshalPatches(actual.Response.Patch)

	podNameSplit := strings.Split(expect.Request.Object.Metadata.Name, "-")
	podIndex := podNameSplit[len(podNameSplit)-1]
	podReplicas := "3" // TODO: use K8s mock api server response to verify the pod replicas

	c := newConstant()
	for i := range patches {
		patch := patches[i]
		if patch.Path == c.AnnotationsPath {
			patchMap := patch.Value.(map[string]interface{})
			actualPodIndex := patchMap["sts-annotator/pod-index"].(string)
			actualPodReplicas := patchMap["sts-annotator/pod-replicas"].(string)
			assert.Equal(t, podIndex, actualPodIndex)
			assert.Equal(t, podReplicas, actualPodReplicas)
		} else if patch.Path == c.PodIndexPath {
			assert.Equal(t, podIndex, patch.Value)
		} else if patch.Path == c.PodReplicasPath {
			assert.Equal(t, podReplicas, patch.Value)
		} else {
			t.Error(
				fmt.Sprintf("JsonPath must be either of %s, %s or %s. But got %s",
					c.AnnotationsPath, c.PodReplicasPath, c.PodIndexPath, patch.Path))
		}
	}
}

func sendRequest(body []byte) (w *httptest.ResponseRecorder) {
	w = httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	router := setupTestRouter()
	router.ServeHTTP(w, req)
	return
}

func unmarshalActual(w *httptest.ResponseRecorder) (actual AdmissionResponse) {
	actual = AdmissionResponse{}
	body, _ := ioutil.ReadAll(w.Body)
	_ = json.Unmarshal(body, &actual)
	return
}

func unmarshalExpected(request []byte) (expect AdmissionRequest) {
	expect = AdmissionRequest{}
	json.Unmarshal(request, &expect)
	return
}

func unmarshalPatches(base64EncodedPatch string) (patches []Patch) {
	patches = []Patch{}
	decodedPatch, _ := base64.StdEncoding.DecodeString(base64EncodedPatch)
	json.Unmarshal(decodedPatch, &patches)
	return
}
