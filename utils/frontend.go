package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func unzip() {
	dst := "dest"
	archive, err := zip.OpenReader("new_frontend.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		// fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			// fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	os.Rename("./dest/ledfx_frontend_v2", "./frontend")

	// cleanup
	if _, err := os.Stat("dest"); err == nil {
		os.RemoveAll("dest")
	}

	defer os.RemoveAll("new_frontend.zip")

}

func GetFrontend() {
	log.Println("Getting latest Frontend")
	resp, err := http.Get("https://github.com/YeonV/LedFx-Frontend-v2/releases/latest/download/ledfx_frontend_v2.zip")
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}
	// Delete old files
	if _, err := os.Stat("frontend"); err == nil {
		os.RemoveAll("frontend")
	}
	defer os.RemoveAll("new_frontend.zip")

	// Create the file
	out, err := os.Create("new_frontend.zip")
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	// Extract frontend
	unzip()
	log.Println("Got latest Frontend")
	fmt.Println("===========================================")
}

func ServeFrontend() {
	serveFrontend := http.FileServer(http.Dir("frontend"))
	fileMatcher := regexp.MustCompile(`\.[a-zA-Z]*$`)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !fileMatcher.MatchString(r.URL.Path) {
			http.ServeFile(w, r, "frontend/index.html")
		} else {
			serveFrontend.ServeHTTP(w, r)
		}
	})
}

func Openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func InitFrontend() {
	fmt.Println("===========================================")
	fmt.Println("          LedFx-Frontend by Blade")
	fmt.Println("    [CTRL]+Click: http://localhost:8080")
	fmt.Println("===========================================")
	SetupRoutes()
	go func() {
		http.ListenAndServe(":8080", nil)
	}()
}
