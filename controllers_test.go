package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-playground/assert/v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var router = setupRouter()
var c = newConstant()
var url = c.V1 + c.Annotator
var method = "POST"

func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", c.V1+c.Healthz, nil)
	router.ServeHTTP(w, req)

	body, _ := ioutil.ReadAll(w.Body)
	expect := Health{Health: true}
	actual := Health{}
	_ = json.Unmarshal(body, &actual)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expect, actual)
}

func TestAnnotateStsPod(t *testing.T) {
	f := readFile("test/admissionRequest.json")
	w := sendRequest(f)
	actual := unmarshalActual(w)
	expect := unmarshalExpected(f)
	assertCommon(t, w.Code, expect, actual)
	assertPatches(t, expect, actual)
}

func TestAnnotateStsPodWithAnnotations(t *testing.T) {
	f := readFile("test/admissionRequestWithAnnotations.json")
	w := sendRequest(f)
	actual := unmarshalActual(w)
	expect := unmarshalExpected(f)
	assertCommon(t, w.Code, expect, actual)
	assertPatches(t, expect, actual)
}

func TestAnnotateNoStsPod(t *testing.T) {
	f := readFile("test/admissionRequestNoSts.json")
	w := sendRequest(f)
	actual := unmarshalActual(w)
	expect := unmarshalExpected(f)
	assertCommon(t, w.Code, expect, actual)
	assert.Equal(t, "", actual.Response.Patch)
	assert.Equal(t, "", actual.Response.PatchType)
}

func TestAnnotateStsPodInvalidReq(t *testing.T) {
	f := readFile("test/admissionRequestInvalid.json")
	w := sendRequest(f)
	actual := unmarshalActual(w)
	assertFailed(t, http.StatusBadRequest, w.Code, actual)
}

func TestAnnotateStsNotFound(t *testing.T) {
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

func assertPatches(t *testing.T, expect AdmissionRequest, actual AdmissionResponse) {
	assert.Equal(t, "jsonpatch", strings.ToLower(actual.Response.PatchType))

	patches := unmarshalPatches(actual.Response.Patch)

	c := newConstant()
	for i := range patches {
		patch := patches[i]
		if patch.Path == c.AnnotationsPath {
			assert.Equal(t, make(map[string]interface{}), patch.Value)
		} else if patch.Path == c.PodIndexPath {
			podNameSplit := strings.Split(expect.Request.Object.Metadata.Name, "-")
			podIndex := podNameSplit[len(podNameSplit)-1]
			assert.Equal(t, podIndex, patch.Value)
		} else if patch.Path == c.PodReplicasPath {
			// TODO: use K8s mock api server response to verify the pod replicas
			assert.NotEqual(t, nil, patch.Value)
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

func readFile(path string) []byte {
	f, _ := os.Open(path)
	defer f.Close()
	var buf bytes.Buffer
	tee := io.TeeReader(f, &buf)
	b, _ := ioutil.ReadAll(tee)
	return b
}
