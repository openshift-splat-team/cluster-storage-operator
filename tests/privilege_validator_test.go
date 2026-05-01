package tests

import (
	"testing"

	"github.com/openshift/cluster-storage-operator/pkg/operator/csidriveroperator/csioperatorclient"
)

// TestRequiredStoragePrivilegesCount verifies that 10-15 storage-specific privileges are defined
func TestRequiredStoragePrivilegesCount(t *testing.T) {
	count := len(csioperatorclient.RequiredStoragePrivileges)

	if count < 10 || count > 15 {
		t.Errorf("Expected 10-15 storage privileges, got %d", count)
	}

	// Ensure no duplicate privileges
	privilegeMap := make(map[string]bool)
	for _, priv := range csioperatorclient.RequiredStoragePrivileges {
		if privilegeMap[priv] {
			t.Errorf("Duplicate privilege found: %s", priv)
		}
		privilegeMap[priv] = true
	}
}

// TestRequiredStoragePrivilegesCategories verifies that required privilege categories are present
func TestRequiredStoragePrivilegesCategories(t *testing.T) {
	requiredCategories := []string{
		"Datastore.AllocateSpace",
		"Datastore.FileManagement",
		"Datastore.Browse",
		"System.Anonymous",
		"System.Read",
		"System.View",
		"VirtualMachine.Config.AddExistingDisk",
		"VirtualMachine.Config.AddNewDisk",
		"VirtualMachine.Config.RemoveDisk",
		"StorageProfile.View",
	}

	// Build map of actual privileges
	privilegeMap := make(map[string]bool)
	for _, priv := range csioperatorclient.RequiredStoragePrivileges {
		privilegeMap[priv] = true
	}

	// Verify key privilege categories are present
	for _, required := range requiredCategories {
		if !privilegeMap[required] {
			t.Errorf("Required privilege not found: %s", required)
		}
	}
}

// TestFormatMissingPrivilegesError verifies error message formatting for missing privileges
func TestFormatMissingPrivilegesError(t *testing.T) {
	vcenter := "vcenter1.example.com"
	missingPrivileges := []string{
		"Datastore.AllocateSpace",
		"VirtualMachine.Config.AddNewDisk",
	}

	err := csioperatorclient.FormatMissingPrivilegesError(vcenter, missingPrivileges)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	errMsg := err.Error()

	// Verify error message includes vCenter FQDN
	if !contains(errMsg, vcenter) {
		t.Errorf("Error message does not include vCenter FQDN: %s", errMsg)
	}

	// Verify error message includes list of missing privileges
	for _, priv := range missingPrivileges {
		if !contains(errMsg, priv) {
			t.Errorf("Error message does not include missing privilege '%s': %s", priv, errMsg)
		}
	}

	// Verify error message includes remediation guidance
	if !contains(errMsg, "grant") && !contains(errMsg, "privileges") {
		t.Errorf("Error message does not include remediation guidance: %s", errMsg)
	}
}

// TestPrivilegeValidationRetry verifies retry logic for privilege validation
func TestPrivilegeValidationRetry(t *testing.T) {
	t.Skip("TODO: Implement test for privilege validation retry logic with mock vSphere client")
	// Test plan:
	// 1. Mock transient vSphere API failures
	// 2. Call ValidatePrivilegesWithRetry()
	// 3. Verify exponential backoff retry behavior
	// 4. Verify max retry limit respected
	// 5. Verify success after transient failure
}

// contains is a helper function for substring checking
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexContains(s, substr))
}

func indexContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
