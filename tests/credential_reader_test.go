package tests

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openshift/cluster-storage-operator/pkg/operator/csidriveroperator/csioperatorclient"
)

// TestReadComponentCredentials verifies that the CSI Driver reads component-specific
// credentials from the openshift-cluster-csi-drivers namespace
func TestReadComponentCredentials(t *testing.T) {
	ctx := context.Background()

	// Create fake secret with component credentials
	componentSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csioperatorclient.ComponentCredentialSecretName,
			Namespace: csioperatorclient.ComponentCredentialNamespace,
		},
		Data: map[string][]byte{
			"vcenter1.example.com.username": []byte("storage@vsphere.local"),
			"vcenter1.example.com.password": []byte("password123"),
		},
	}

	// Create fake Kubernetes client
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(componentSecret).Build()

	// Initialize credential reader
	reader := csioperatorclient.NewCredentialReader(fakeClient)

	// Call GetCredentialsForVCenter
	cred, err := reader.GetCredentialsForVCenter(ctx, "vcenter1.example.com")
	if err != nil {
		t.Fatalf("GetCredentialsForVCenter failed: %v", err)
	}

	// Verify credentials retrieved from openshift-cluster-csi-drivers namespace
	if cred.Server != "vcenter1.example.com" {
		t.Errorf("Expected server 'vcenter1.example.com', got '%s'", cred.Server)
	}
	if cred.Username != "storage@vsphere.local" {
		t.Errorf("Expected username 'storage@vsphere.local', got '%s'", cred.Username)
	}
	if cred.Password != "password123" {
		t.Errorf("Expected password 'password123', got '%s'", cred.Password)
	}
}

// TestComponentCredentialFormat verifies that FQDN-keyed credential format is correctly parsed
func TestComponentCredentialFormat(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		secret      *corev1.Secret
		vcenterFQDN string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid FQDN-keyed credentials",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      csioperatorclient.ComponentCredentialSecretName,
					Namespace: csioperatorclient.ComponentCredentialNamespace,
				},
				Data: map[string][]byte{
					"vcenter.example.com.username": []byte("user"),
					"vcenter.example.com.password": []byte("pass"),
				},
			},
			vcenterFQDN: "vcenter.example.com",
			expectError: false,
		},
		{
			name: "missing username",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      csioperatorclient.ComponentCredentialSecretName,
					Namespace: csioperatorclient.ComponentCredentialNamespace,
				},
				Data: map[string][]byte{
					"vcenter.example.com.password": []byte("pass"),
				},
			},
			vcenterFQDN: "vcenter.example.com",
			expectError: true,
			errorMsg:    "missing credentials",
		},
		{
			name: "missing password",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      csioperatorclient.ComponentCredentialSecretName,
					Namespace: csioperatorclient.ComponentCredentialNamespace,
				},
				Data: map[string][]byte{
					"vcenter.example.com.username": []byte("user"),
				},
			},
			vcenterFQDN: "vcenter.example.com",
			expectError: true,
			errorMsg:    "missing credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.secret).Build()
			reader := csioperatorclient.NewCredentialReader(fakeClient)

			_, err := reader.GetCredentialsForVCenter(ctx, tt.vcenterFQDN)

			if tt.expectError && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestCredentialFallback verifies fallback to shared credentials when component credentials not found
func TestCredentialFallback(t *testing.T) {
	ctx := context.Background()

	// Create fake client WITHOUT vsphere-storage-creds secret
	// Create vsphere-cloud-credentials secret in openshift-config namespace
	sharedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csioperatorclient.SharedCredentialSecretName,
			Namespace: csioperatorclient.SharedCredentialNamespace,
		},
		Data: map[string][]byte{
			"vcenter1.example.com.username": []byte("shared@vsphere.local"),
			"vcenter1.example.com.password": []byte("shared-password"),
		},
	}

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sharedSecret).Build()
	reader := csioperatorclient.NewCredentialReader(fakeClient)

	// Call GetCredentialsForVCenter
	cred, err := reader.GetCredentialsForVCenter(ctx, "vcenter1.example.com")
	if err != nil {
		t.Fatalf("GetCredentialsForVCenter failed: %v", err)
	}

	// Verify fallback to shared credentials
	if cred.Username != "shared@vsphere.local" {
		t.Errorf("Expected fallback username 'shared@vsphere.local', got '%s'", cred.Username)
	}
	if cred.Password != "shared-password" {
		t.Errorf("Expected fallback password 'shared-password', got '%s'", cred.Password)
	}
}

// TestGetAllVCentersFromSecret verifies extraction of all vCenter FQDNs from secret
func TestGetAllVCentersFromSecret(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csioperatorclient.ComponentCredentialSecretName,
			Namespace: csioperatorclient.ComponentCredentialNamespace,
		},
		Data: map[string][]byte{
			"vcenter1.example.com.username": []byte("user1"),
			"vcenter1.example.com.password": []byte("pass1"),
			"vcenter2.example.com.username": []byte("user2"),
			"vcenter2.example.com.password": []byte("pass2"),
			"vcenter3.example.com.username": []byte("user3"),
			"vcenter3.example.com.password": []byte("pass3"),
		},
	}

	// Call GetAllVCentersFromSecret
	vcenters := csioperatorclient.GetAllVCentersFromSecret(secret)

	// Verify all FQDNs extracted correctly
	if len(vcenters) != 3 {
		t.Fatalf("Expected 3 vCenters, got %d", len(vcenters))
	}

	// Verify no duplicates
	vcenterMap := make(map[string]bool)
	for _, vcenter := range vcenters {
		if vcenterMap[vcenter] {
			t.Errorf("Duplicate vCenter FQDN: %s", vcenter)
		}
		vcenterMap[vcenter] = true
	}

	// Verify expected vCenters are present
	expectedVCenters := []string{
		"vcenter1.example.com",
		"vcenter2.example.com",
		"vcenter3.example.com",
	}

	for _, expected := range expectedVCenters {
		if !vcenterMap[expected] {
			t.Errorf("Expected vCenter FQDN '%s' not found", expected)
		}
	}
}
