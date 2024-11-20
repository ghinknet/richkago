package richkago

import "testing"

// TestNewController test example that builds a new controller
func TestNewController(t *testing.T) {
	controller := NewController()
	if controller == nil {
		t.Error("controller is nil")
		return
	}
}

// TestController_UpdateProgress test example that updates progress
func TestController_UpdateProgress(t *testing.T) {
	controller := NewController()
	controller.totalSize = 1000

	controller.UpdateProgress(100, "1")
	if controller.Progress() != 10 {
		t.Error("progress is wrong", controller.Progress())
		return
	}

	if controller.Status() != 1 {
		t.Error("status is wrong", controller.Status())
		return
	}
}

// TestController_Pause test example that pause a progress
func TestController_Pause(t *testing.T) {
	controller := NewController()
	controller.totalSize = 1000

	controller.UpdateProgress(100, "1")
	if controller.Progress() != 10 {
		t.Error("progress is wrong", controller.Progress())
		return
	}

	controller.Pause()
	if controller.Status() != -2 {
		t.Error("status is wrong", controller.Status())
		return
	}
}

// TestController_Unpause test example that unpause a progress
func TestController_Unpause(t *testing.T) {
	controller := NewController()
	controller.totalSize = 1000

	controller.UpdateProgress(100, "1")
	if controller.Progress() != 10 {
		t.Error("progress is wrong", controller.Progress())
		return
	}

	controller.Pause()
	if controller.Status() != -2 {
		t.Error("status is wrong", controller.Status())
		return
	}

	controller.Unpause()
	if controller.Status() != 1 {
		t.Error("status is wrong", controller.Status())
		return
	}
}
