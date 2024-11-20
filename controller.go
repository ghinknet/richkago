package richkago

import "sync"

type Controller struct {
	paused           bool
	excepted         bool
	totalSize        int64
	downloadedSize   int64
	downloadedSlices map[string]int64
	mu               sync.Mutex
}

// NewController build a new controller
func NewController() *Controller {
	return &Controller{
		downloadedSlices: make(map[string]int64),
	}
}

// UpdateProgress update progress into controller
func (c *Controller) UpdateProgress(size int64, chunkID string) {
	// Get lock
	c.mu.Lock()
	defer c.mu.Unlock()

	if chunkID == "" && len(c.downloadedSlices) == 0 {
		// Init variable
		c.downloadedSize = size
	} else {
		// Update progress
		c.downloadedSlices[chunkID] = size
		c.downloadedSize = 0
		// Sum up
		for _, v := range c.downloadedSlices {
			c.downloadedSize += v
		}
	}
}

// Pause pause a progress
func (c *Controller) Pause() {
	c.paused = true
}

// Unpause unpause a progress
func (c *Controller) Unpause() {
	c.paused = false
}

// Status gets a status of a controller
func (c *Controller) Status() int {
	if c.downloadedSize == 0 && !c.excepted {
		return -1 // Not started
	} else if c.paused {
		return -2 // Paused
	} else if c.excepted {
		return -3 // Excepted
	} else if c.downloadedSize == c.totalSize {
		return 0 // Done
	} else {
		return 1 // Downloading
	}
}

// Progress gets progress of a controller
func (c *Controller) Progress() float64 {
	if c.totalSize == 0 {
		return -1
	}

	return float64(c.downloadedSize) / float64(c.totalSize) * 100
}
