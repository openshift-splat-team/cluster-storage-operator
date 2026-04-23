/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package csioperatorclient

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ComponentCredentialSecretName is the name of the secret containing component-specific vSphere credentials
	ComponentCredentialSecretName = "vsphere-storage-creds"

	// ComponentCredentialNamespace is the namespace where component credentials are stored
	ComponentCredentialNamespace = "openshift-cluster-csi-drivers"

	// SharedCredentialSecretName is the fallback secret name for passthrough mode
	SharedCredentialSecretName = "vsphere-cloud-credentials"

	// SharedCredentialNamespace is the namespace for shared credentials
	SharedCredentialNamespace = "openshift-config"
)

// CredentialReader reads vSphere credentials from Kubernetes secrets
type CredentialReader struct {
	client client.Client
}

// NewCredentialReader creates a new CredentialReader
func NewCredentialReader(c client.Client) *CredentialReader {
	return &CredentialReader{
		client: c,
	}
}

// VCenterCredential contains authentication details for a vCenter
type VCenterCredential struct {
	Server   string
	Username string
	Password string
}

// GetCredentialsForVCenter retrieves credentials for a specific vCenter FQDN
// It first attempts to read component-specific credentials from openshift-cluster-csi-drivers namespace,
// then falls back to shared credentials if component credentials are not found
func (cr *CredentialReader) GetCredentialsForVCenter(ctx context.Context, vcenterFQDN string) (*VCenterCredential, error) {
	klog.V(4).Infof("Fetching credentials for vCenter: %s", vcenterFQDN)

	// Try component-specific credentials first
	cred, err := cr.getComponentCredentials(ctx, vcenterFQDN)
	if err == nil {
		klog.V(4).Infof("Using component-specific credentials for vCenter: %s", vcenterFQDN)
		return cred, nil
	}

	klog.V(4).Infof("Component credentials not found for %s, falling back to shared credentials: %v", vcenterFQDN, err)

	// Fall back to shared credentials
	return cr.getSharedCredentials(ctx, vcenterFQDN)
}

// getComponentCredentials reads component-specific credentials from the openshift-cluster-csi-drivers namespace
func (cr *CredentialReader) getComponentCredentials(ctx context.Context, vcenterFQDN string) (*VCenterCredential, error) {
	secret := &corev1.Secret{}
	err := cr.client.Get(ctx, types.NamespacedName{
		Namespace: ComponentCredentialNamespace,
		Name:      ComponentCredentialSecretName,
	}, secret)

	if err != nil {
		return nil, fmt.Errorf("failed to read component credential secret: %w", err)
	}

	return extractCredentialFromSecret(secret, vcenterFQDN)
}

// getSharedCredentials reads shared credentials from the openshift-config namespace
func (cr *CredentialReader) getSharedCredentials(ctx context.Context, vcenterFQDN string) (*VCenterCredential, error) {
	secret := &corev1.Secret{}
	err := cr.client.Get(ctx, types.NamespacedName{
		Namespace: SharedCredentialNamespace,
		Name:      SharedCredentialSecretName,
	}, secret)

	if err != nil {
		return nil, fmt.Errorf("failed to read shared credential secret: %w", err)
	}

	return extractCredentialFromSecret(secret, vcenterFQDN)
}

// extractCredentialFromSecret extracts vCenter credentials from a secret
// Secret format for component credentials:
//   vcenter.example.com.username: "user@vsphere.local"
//   vcenter.example.com.password: "password"
func extractCredentialFromSecret(secret *corev1.Secret, vcenterFQDN string) (*VCenterCredential, error) {
	usernameKey := vcenterFQDN + ".username"
	passwordKey := vcenterFQDN + ".password"

	username, usernameFound := secret.Data[usernameKey]
	password, passwordFound := secret.Data[passwordKey]

	if !usernameFound || !passwordFound {
		// Fallback: try generic username/password keys (for shared credential secret)
		username, usernameFound = secret.Data["username"]
		password, passwordFound = secret.Data["password"]

		if !usernameFound || !passwordFound {
			return nil, fmt.Errorf("missing credentials for vCenter %s in secret %s/%s",
				vcenterFQDN, secret.Namespace, secret.Name)
		}
	}

	return &VCenterCredential{
		Server:   vcenterFQDN,
		Username: string(username),
		Password: string(password),
	}, nil
}

// ValidateSecretFormat checks if the secret has the expected format for component credentials
func ValidateSecretFormat(secret *corev1.Secret) error {
	if secret == nil {
		return fmt.Errorf("secret is nil")
	}

	if secret.Data == nil || len(secret.Data) == 0 {
		return fmt.Errorf("secret %s/%s has no data", secret.Namespace, secret.Name)
	}

	// Check if at least one vCenter credential pair exists
	hasValidCredential := false
	for key := range secret.Data {
		if strings.HasSuffix(key, ".username") {
			passwordKey := strings.TrimSuffix(key, ".username") + ".password"
			if _, ok := secret.Data[passwordKey]; ok {
				hasValidCredential = true
				break
			}
		}
	}

	// Also check for generic username/password (shared credential format)
	if _, usernameOk := secret.Data["username"]; usernameOk {
		if _, passwordOk := secret.Data["password"]; passwordOk {
			hasValidCredential = true
		}
	}

	if !hasValidCredential {
		return fmt.Errorf("secret %s/%s does not contain valid credential pairs", secret.Namespace, secret.Name)
	}

	return nil
}

// GetAllVCentersFromSecret extracts all vCenter FQDNs from a credential secret
func GetAllVCentersFromSecret(secret *corev1.Secret) []string {
	vcenters := make(map[string]bool)

	for key := range secret.Data {
		if strings.HasSuffix(key, ".username") {
			vcenterFQDN := strings.TrimSuffix(key, ".username")
			vcenters[vcenterFQDN] = true
		}
	}

	result := make([]string, 0, len(vcenters))
	for vcenter := range vcenters {
		result = append(result, vcenter)
	}

	return result
}
