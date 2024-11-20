# Richkago - GoLang Download Engine

#### Richka (From Ukrainian: Рiчка) means river, stands for the download speed of Richka Engine

#### This project is only active on GitHub (https://github.com/ghinknet/richkago), repositories on any other platform are mirror
#### We will not process PRs or Issues outside of GitHub

#### Richka for python is here (https://github.com/ghinknet/richka)

## Usage

`import "github.com/ghinknet/richkago"` and run script in your code, for example:

```
package main

import (
	"fmt"
	"github.com/ghinknet/richkago"
	"time"
)

func main() {
	controller := richkago.NewController()

	go func() {
		// Start download
		timeUsed, fileSize, err := richkago.Download("https://mirrors.tuna.tsinghua.edu.cn/raspberry-pi-os-images/raspios_lite_arm64/images/raspios_lite_arm64-2024-11-19/2024-11-19-raspios-bookworm-arm64-lite.img.xz", "2024-11-19-raspios-bookworm-arm64-lite3.img.xz", controller)
		if err != nil {
			fmt.Println("Download failed:", err)
			return
		}

		// Print result
		fmt.Printf("Time used: %.2f seconds\n", timeUsed)
		fmt.Printf("Speed: %.2f MiB/s\n", float64(fileSize)/timeUsed/1024/1024)
	}()

	// Monitor progress
	for controller.Status() != 0 && controller.Status() != -3 {
		if controller.Status() == 1 {
			fmt.Printf("Download Progress: %.2f%%\r", controller.Progress())
		}
		time.Sleep(100 * time.Millisecond)
	}
	if controller.Status() == 0 {
		fmt.Println("Download completed.")
	}
}
```
Then you'll get a file from Internet :D.