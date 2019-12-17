package detector

import (
	"testing"
)

func TestStateManager_Test_(t *testing.T) {
	stateManager := GetStateManager()
	stateManagerTwo := GetStateManager()

	stateManager.SetState("Test", "string")

	if stateManager.GetState("NotSet") != nil {
		t.Fatalf("Expected nil, got '%s'", stateManager.GetState("NotSet"))
	}

	if stateManager.GetState("Test") != "string" {
		t.Fatalf("Expected 'string' got '%s'", stateManager.GetState("Test"))
	}

	if stateManager != stateManagerTwo {
		t.Fatalf("Expected objects to be equal")
	}
}
