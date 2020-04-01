package main

type Config struct {
	Server struct {
		Mode string `yaml:"mode"`
		Log  struct {
			Level string `yaml:"level"`
		} `yaml:"log"`
		Port int `yaml:"port"`
		Tls  struct {
			Cert string `yaml:"cert"`
			Key  string `yaml:"key"`
		} `yaml:"tls"`
	} `yaml:"server"`
	K8s struct {
		URL   string `yaml:"url"`
		Token string `yaml:"token"`
		Tls   struct {
			CaCert   string      `yaml:"caCert"`
			CaBundle interface{} `yaml:"caBundle"`
		} `yaml:"tls"`
	} `yaml:"k8s"`
}

type Health struct {
	Health bool `json:"health"`
}

type AdmissionRequest struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Request    struct {
		UID       string `json:"uid"`
		Namespace string `json:"namespace"`
		Operation string `json:"operation"`
		Object    struct {
			Metadata struct {
				Name            string           `json:"name"`
				GenerateName    string           `json:"generateName"`
				Namespace       string           `json:"namespace"`
				Annotations     interface{}      `json:"annotations"`
				OwnerReferences []OwnerReference `json:"ownerReferences"`
			} `json:"metadata"`
		} `json:"object"`
	} `json:"request"`
}

type OwnerReference struct {
	APIVersion         string `json:"apiVersion"`
	Kind               string `json:"kind"`
	Name               string `json:"name"`
	UID                string `json:"uid"`
	Controller         bool   `json:"controller"`
	BlockOwnerDeletion bool   `json:"blockOwnerDeletion"`
}

type AdmissionResponse struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Response   Response `json:"response"`
}

type Response struct {
	UID       string `json:"uid"`
	Allowed   bool   `json:"allowed"`
	PatchType string `json:"patchType,omitempty"`
	Patch     string `json:"patch,omitempty"`
}

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type Statefulset struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Replicas int `json:"replicas"`
	} `json:"spec"`
}

func newPatch() (p Patch) {
	p = Patch{}
	p.Op = "add"
	return
}

type Constant struct {
	// API Path
	V1        string
	Healthz   string
	Annotator string
	// Annotation Path
	AnnotationsPath string
	PodReplicasPath string
	PodIndexPath    string
}

func newConstant() (c Constant) {
	c = Constant{}
	c.Healthz = "/healthz"
	c.V1 = "/v1"
	c.Annotator = "/sts/pod/annotation"
	c.AnnotationsPath = "/metadata/annotations"
	c.PodIndexPath = c.AnnotationsPath + "/sts-annotator~1pod-index"
	c.PodReplicasPath = c.AnnotationsPath + "/metadata/annotations/sts-annotator~1pod-replicas"
	return
}
