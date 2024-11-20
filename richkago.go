package richkago

import (
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	version        = "Alpha/0.0.1"
	userAgent      = "Richkago" + version
	coroutineLimit = 10
	sliceThreshold = 10 * 1024 * 1024 // 10 MiB
	timeout        = 3 * time.Second
	retryTimes     = 5
	chunkSize      = 102400
)

// readWithTimeout read data from source with timeout limit
func readWithTimeout(reader io.Reader, buffer []byte, controller *Controller, file *os.File, downloaded *int64, chunkID string) error {
	// Make a error chan
	ch := make(chan error, 1)

	go func() {
		// Read data
		n, err := reader.Read(buffer)
		if n > 0 {
			// Write to file
			_, err = file.Write(buffer[:n])
			if err != nil {
				ch <- err
				return
			}
			// Calc amount of downloaded data
			*downloaded += int64(n)
		}
		ch <- err
	}()

	// Error and timeout handler
	select {
	case err := <-ch:
		// Update amount of downloaded data
		if controller != nil {
			controller.UpdateProgress(*downloaded, chunkID)
		}

		return err
	case <-time.After(timeout):
		return errors.New("timeout while reading data")
	}
}

// downloadRange download a file with many slices
func downloadRange(client *http.Client, url string, start, end int64, destination string, controller *Controller) error {
	// Build request header
	headers := map[string]string{
		"User-Agent": userAgent,
		"Range":      fmt.Sprintf("bytes=%d-%d", start, end),
	}

	retries := retryTimes
	var err error

	for retries > 0 {
		// Get header info
		var req *http.Request
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			retries--
			time.Sleep(1 * time.Second)
			continue
		}
		for k, v := range headers {
			req.Header.Add(k, v)
		}

		// Read http header
		var resp *http.Response
		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusPartialContent {
			retries--
			time.Sleep(1 * time.Second)
			continue
		}

		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				panic(err)
			}
		}(resp.Body)

		// Pre-create file
		file, _ := os.OpenFile(destination, os.O_WRONLY, 0644)
		defer func(file *os.File) {
			err = file.Close()
			if err != nil {
				panic(err)
			}
		}(file)
		_, err = file.Seek(start, io.SeekStart)
		if err != nil {
			retries--
			time.Sleep(1 * time.Second)
			continue
		}

		// Create goroutines to download
		buffer := make([]byte, chunkSize)
		var downloaded int64
		for {
			// Controller pin
			if controller.paused {
				time.Sleep(1 * time.Second)
				continue
			}

			// Start read stream
			err = readWithTimeout(resp.Body, buffer, controller, file, &downloaded, fmt.Sprintf("%d-%d", start, end))
			if err == io.EOF {
				break
			} else if err != nil {
				break
			}
		}

		// Error handler
		if err != nil && err != io.EOF {
			retries--
			time.Sleep(1 * time.Second)
			continue
		}

		return nil
	}

	return err
}

// downloadSingle download a file directly
func downloadSingle(client *http.Client, url, destination string, controller *Controller) error {
	// Build request header
	headers := map[string]string{
		"User-Agent": userAgent,
	}
	retries := retryTimes

	for retries > 0 {
		// Get header info
		req, _ := http.NewRequest("GET", url, nil)
		for k, v := range headers {
			req.Header.Add(k, v)
		}

		// Read http header
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			retries--
			time.Sleep(1 * time.Second)
			continue
		}

		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				panic(err)
			}
		}(resp.Body)

		// Pre-create file
		file, _ := os.OpenFile(destination, os.O_WRONLY, 0644)
		defer func(file *os.File) {
			err = file.Close()
			if err != nil {
				panic(err)
			}
		}(file)

		// Start downloading
		buffer := make([]byte, chunkSize)
		var downloaded int64
		for {
			// Controller pin
			if controller.paused {
				time.Sleep(1 * time.Second)
				continue
			}

			// Start read stream
			var n int
			n, err = resp.Body.Read(buffer)
			if err != nil {
				break
			}
			if n > 0 {
				_, err = file.Write(buffer[:n])
				if err != nil {
					return err
				}
				downloaded += int64(n)
				controller.UpdateProgress(downloaded, "")
			}
		}

		// Error handler
		if err != nil {
			retries--
			time.Sleep(1 * time.Second)
			continue
		} else {
			return err
		}
	}

	return nil
}

// Download main download task creator
func Download(url, destination string, controller *Controller) (float64, int64, error) {
	// Create http client
	client := &http.Client{}

	// Get header info
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		controller.excepted = true
		return 0, 0, err
	}
	req.Header.Add("User-Agent", userAgent)

	// Read http header
	resp, err := client.Do(req)
	if err != nil {
		controller.excepted = true
		return 0, 0, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	sizeStr := resp.Header.Get("Content-Length")
	fileSize, _ := strconv.ParseInt(sizeStr, 10, 64)
	controller.totalSize = fileSize

	// Calc slices size
	if fileSize <= sliceThreshold {
		file, _ := os.Create(destination)
		err = file.Truncate(fileSize)
		if err != nil {
			controller.excepted = true
			return 0, 0, err
		}
		err = file.Close()
		if err != nil {
			controller.excepted = true
			return 0, 0, err
		}

		// Too small
		startTime := time.Now()
		err = downloadSingle(client, url, destination, controller)
		if err != nil {
			controller.excepted = true
			return 0, 0, err
		}
		endTime := time.Now()
		return endTime.Sub(startTime).Seconds(), fileSize, nil
	}

	// Pre-create file
	partSize := fileSize / coroutineLimit
	file, _ := os.Create(destination)
	err = file.Truncate(fileSize)
	if err != nil {
		controller.excepted = true
		return 0, 0, err
	}
	err = file.Close()
	if err != nil {
		controller.excepted = true
		return 0, 0, err
	}

	// Start download goroutines
	group := new(errgroup.Group)
	for i := 0; i < coroutineLimit; i++ {
		start := int64(i) * partSize
		end := start + partSize - 1
		if i == coroutineLimit-1 {
			end = fileSize - 1
		}
		group.Go(func() error {
			err = downloadRange(client, url, start, end, destination, controller)
			return err
		})
	}

	// Start all tasks
	startTime := time.Now()
	if err = group.Wait(); err != nil {
		controller.excepted = true
		return 0, 0, err
	}
	endTime := time.Now()

	return endTime.Sub(startTime).Seconds(), fileSize, nil
}
