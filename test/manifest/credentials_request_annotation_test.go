package manifest_test

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const vsphereComponentAnnotation = "cloudcredential.openshift.io/vsphere-component"

type credentialsRequest struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name        string            `yaml:"name"`
		Annotations map[string]string `yaml:"annotations"`
	} `yaml:"metadata"`
}

func parseCredentialsRequests(t *testing.T, path string) []credentialsRequest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out []credentialsRequest
	for _, doc := range strings.Split(string(data), "\n---") {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}
		var cr credentialsRequest
		if err := yaml.Unmarshal([]byte(doc), &cr); err != nil {
			t.Fatalf("unmarshal %s: %v", path, err)
		}
		if cr.Kind == "CredentialsRequest" {
			out = append(out, cr)
		}
	}
	return out
}

func findCR(crs []credentialsRequest, name string) (credentialsRequest, bool) {
	for _, cr := range crs {
		if cr.Metadata.Name == name {
			return cr, true
		}
	}
	return credentialsRequest{}, false
}

func TestCSOCSIVSphereCredentialsRequestAnnotation(t *testing.T) {
	path := "../../manifests/03_credentials_request_vsphere_csi.yaml"
	crs := parseCredentialsRequests(t, path)

	cr, ok := findCR(crs, "openshift-vmware-vsphere-csi-driver-operator")
	if !ok {
		t.Fatal("CredentialsRequest 'openshift-vmware-vsphere-csi-driver-operator' not found in manifest")
	}

	got, present := cr.Metadata.Annotations[vsphereComponentAnnotation]
	if !present {
		t.Fatalf("annotation %q missing from openshift-vmware-vsphere-csi-driver-operator", vsphereComponentAnnotation)
	}
	if got != "csiDriver" {
		t.Errorf("annotation value: got %q, want %q", got, "csiDriver")
	}
}

func TestCSODetectorVSphereCredentialsRequestAnnotation(t *testing.T) {
	path := "../../manifests/03_credentials_request_vsphere_detector.yaml"
	crs := parseCredentialsRequests(t, path)

	cr, ok := findCR(crs, "openshift-vsphere-problem-detector")
	if !ok {
		t.Fatal("CredentialsRequest 'openshift-vsphere-problem-detector' not found in manifest")
	}

	got, present := cr.Metadata.Annotations[vsphereComponentAnnotation]
	if !present {
		t.Fatalf("annotation %q missing from openshift-vsphere-problem-detector", vsphereComponentAnnotation)
	}
	// Value must match CCO credential_distribution.go componentSecretName() case "vsphereProblemDetector"
	if got != "vsphereProblemDetector" {
		t.Errorf("annotation value: got %q, want %q", got, "vsphereProblemDetector")
	}
}

func TestAnnotationValuesAreNotEmpty(t *testing.T) {
	type tc struct {
		path   string
		crName string
	}
	cases := []tc{
		{
			path:   "../../manifests/03_credentials_request_vsphere_csi.yaml",
			crName: "openshift-vmware-vsphere-csi-driver-operator",
		},
		{
			path:   "../../manifests/03_credentials_request_vsphere_detector.yaml",
			crName: "openshift-vsphere-problem-detector",
		},
	}
	for _, c := range cases {
		crs := parseCredentialsRequests(t, c.path)
		cr, ok := findCR(crs, c.crName)
		if !ok {
			t.Errorf("CR %q not found in %s", c.crName, c.path)
			continue
		}
		val := cr.Metadata.Annotations[vsphereComponentAnnotation]
		if strings.TrimSpace(val) == "" {
			t.Errorf("CR %q: annotation %q must not be empty", c.crName, vsphereComponentAnnotation)
		}
	}
}
