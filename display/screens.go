package display

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"time"
	"github.com/shirou/gopsutil/v4/mem"

	humanize "github.com/dustin/go-humanize"
	"cloudkey/images"
	"github.com/jnovack/cloudkey/src/network"
	"github.com/jnovack/speedtest"


	linuxproc "github.com/c9s/goprocinfo/linux"
)

func buildNetwork(i int, demo bool) {
	screen := screens[i]
	hostname := "Simons cloudkey"
	lan := "192.168.11.13"
	wan := "203.0.113.32"

	draw.Draw(screen, screen.Bounds(), image.Black, image.ZP, draw.Src)
	draw.Draw(screen, image.Rect(2, 2, 2+16, 2+16), images.Load("host"), image.ZP, draw.Src)
	draw.Draw(screen, image.Rect(2, 22, 2+16, 22+16), images.Load("network"), image.ZP, draw.Src)
	draw.Draw(screen, image.Rect(2, 42, 2+16, 42+16), images.Load("internet"), image.ZP, draw.Src)

	// Loop Every Hour
	go func() {
		for {
			if !demo {
				hostname, _ = os.Hostname()
			}
			write(screen, hostname, 22, 1, 12, "lato-regular")

			if !demo {
				lan, _ = network.LANIP()
			}
			write(screen, lan, 22, 21, 12, "lato-regular")

			if !demo {
				wan, _ = network.WANIP()
			}
			write(screen, wan, 22, 41, 12, "lato-regular")

			time.Sleep(59 * time.Minute)
		}
	}()
}

func buildSpeedTest(i int, demo bool) {
	dmsg := "calculating..."
	umsg := "calculating..."
	tmsg := "in progress"

	download := make(chan int)
	upload := make(chan int)
	lastcheck := time.Now()

	screen := screens[i]

	draw.Draw(screen, screen.Bounds(), image.Black, image.ZP, draw.Src)
	draw.Draw(screen, image.Rect(2, 2, 2+16, 2+16), images.Load("download"), image.ZP, draw.Src)
	draw.Draw(screen, image.Rect(2, 22, 2+16, 22+16), images.Load("upload"), image.ZP, draw.Src)
	draw.Draw(screen, image.Rect(2, 42, 2+16, 42+16), images.Load("clock"), image.ZP, draw.Src)

	if demo {
		dmsg = "86.1 Mb/s"
		umsg = "43.9 Mb/s"
		tmsg = "25 minutes ago"
		write(screen, dmsg, 22, 1, 12, "lato-regular")
		write(screen, umsg, 22, 21, 12, "lato-regular")
		write(screen, tmsg, 22, 41, 12, "lato-regular")
	} else {

		client := speedtest.NewClient(&speedtest.Opts{})

		// Loop every 10 Seconds
		go func() {
			for {
				tmsg = fmt.Sprintf("%s", humanize.Time(lastcheck))
				draw.Draw(screen, image.Rect(20, 0, 160, 60), image.Black, image.ZP, draw.Src)
				write(screen, dmsg, 22, 1, 12, "lato-regular")
				write(screen, umsg, 22, 21, 12, "lato-regular")
				write(screen, tmsg, 22, 41, 12, "lato-regular")
				time.Sleep(10 * time.Second)
			}
		}()

		// Loop Every Hour
		go func() {
			for {
				myLeds.LED("blue").Blink(128, 500, 500)

				server := client.SelectServer(&speedtest.Opts{})

				fmt.Printf("Hosted by %s (%s) [%.2f km]: %d ms\n",
					server.Sponsor,
					server.Name,
					server.Distance,
					server.Latency/time.Millisecond)

				go func() { download <- server.DownloadSpeed() }()

			Download:
				for {
					select {
					case dlspeed := <-download:
						dmsg = fmt.Sprintf("%.2f Mb", float64(dlspeed)/(1<<17))
						break Download
					}
				}

				go func() { upload <- server.UploadSpeed() }()

			Upload:
				for {
					select {
					case ulspeed := <-upload:
						umsg = fmt.Sprintf("%.2f Mb", float64(ulspeed)/(1<<17))
						break Upload
					}
				}

				lastcheck = time.Now()
				fmt.Printf("Download: %s / Upload: %s\n", dmsg, umsg)
				myLeds.LED("blue").On()
				time.Sleep(59 * time.Minute)
			}
		}()
	}
}

func buildSystemStats(i int, demo bool) {
	

	screen := screens[i]

	// Clear the screen
	draw.Draw(screen, screen.Bounds(), image.Black, image.ZP, draw.Src)

	// Draw static labels for CPU and RAM
	draw.Draw(screen, image.Rect(2, 2, 2+16, 22+16), images.Load("ram"), image.ZP, draw.Src)
	draw.Draw(screen, image.Rect(2, 22, 2+16, 22+16), images.Load("cpu"), image.ZP, draw.Src)

	// Loop to update stats periodically
	go func() {
			for {
				v, _ := mem.VirtualMemory()
				used := float64(v.Used)/(1024*1024*1024)
				total := float64(v.Total)/(1024*1024*1024)
				usedPercent := v.UsedPercent
				
				ramInfo := fmt.Sprintf(" %.1f/%.1fGB %.1f%%", used, total, usedPercent)
				
				cpuUsage, _ := getCPUUsagePerCore()
				cpuInfo := fmt.Sprintf(" %.1f%%", cpuUsage)

				// fmt.Println("Used:", used)
				// fmt.Println("Total:", total)
				fmt.Println("CPU Usage:", cpuInfo)

				write(screen, ramInfo, 22, 1, 12, "lato-regular")
				write(screen, cpuInfo, 22, 21, 12, "lato-regular")

				time.Sleep(5 * time.Second)
			}
	}()
}


func getCPUUsagePerCore() (float64, error) {
	// Read CPU stats
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		return 0, err
	}

	// Loop through all cores and calculate the usage
	var totalCPUUsage uint64
	var totalCPUTime uint64
	for _ , stats := range stat.CPUStats {
		// Extract stats for each core
		user := stats.User
		system := stats.System
		idle := stats.Idle
		IOWait := stats.IOWait

		// Calculate total time spent (user + system + idle + IOWait)
		total := user + system + idle + IOWait

		// Calculate the total active time (user + system + IOWait)
		active := user + system + IOWait

		// Accumulate totals
		totalCPUUsage += active
		totalCPUTime += total
	}

	// Calculate the total CPU usage as a percentage
	if totalCPUTime == 0 {
		return 0, nil // Avoid division by zero
	}

	usagePercentage := (float64(totalCPUUsage) / float64(totalCPUTime)) * 100
	return usagePercentage, nil
}

