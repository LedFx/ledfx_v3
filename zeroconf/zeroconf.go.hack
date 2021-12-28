package zeroconf

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/oleksandr/bonjour"
)

func ScanZeroconf() error {
	errch := make(chan error, 1)
	cmd := exec.Command("dns-sd", "-B _wled")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	go func() {
		errch <- cmd.Wait()
	}()

	go func() {
		for _, char := range "|/-\\" {
			fmt.Printf("\r%s...%c", "NEW Scanning", char)
			time.Sleep(100 * time.Millisecond)
		}
		scanner := bufio.NewScanner(stdout)
		fmt.Println("")
		for scanner.Scan() {
			line := scanner.Text()
			log.Println(line)
		}
	}()

	select {
	case <-time.After(time.Second * 1):
		log.Println("NEW Scanning done!")

		log.Println("OLD Scanning")
		resolver, err := bonjour.NewResolver(nil)
		if err != nil {
			log.Println("Failed to initialize resolver:", err.Error())
			os.Exit(1)
		}

		results := make(chan *bonjour.ServiceEntry)

		go func(results chan *bonjour.ServiceEntry) {
			for e := range results {
				fmt.Printf("Found WLED: %s\n", e.Instance)
				// exitCh <- true
				// time.Sleep(1e9)
				// os.Exit(0)
			}
		}(results)

		err = resolver.Browse("_wled._tcp.", "local.", results)
		if err != nil {
			log.Println("Failed to browse:", err.Error())
		}
		return nil
	case err := <-errch:
		if err != nil {
			log.Println("traceroute failed:", err)
		}
	}
	return nil
}
