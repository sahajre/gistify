package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type gistinfo struct {
	ID  string `json:",omitempty"`
	URL string `json:",omitempty"`

	Filename string
	Lastmod  int64
	Public   bool
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		*files = append(*files, path)
		return nil
	}
}

// const version = "0.0.1"
// const usage = `usage:
// gistify [-dry-run] [pattern]
// Note: Set GISTFY_TOKEN environment variable`
const gistifyFile = ".gistify"

func main() {
	token := os.Getenv("GISTIFY_TOKEN")
	if len(token) == 0 {
		log.Fatal("ERROR: could not find GISTIFY_TOKEN environment variable set")
	}

	searchDir := "."
	var files []string
	err := filepath.Walk(searchDir, visit(&files))
	if err != nil {
		log.Fatalf("ERROR: could not search files in dir %s: %v", searchDir, err)
	}

	srchPtrn := os.Args[1] //Example: ".*.go"
	r, err := regexp.Compile(srchPtrn)
	if err != nil {
		log.Fatalf("could not compile given regex %s: %v", srchPtrn, err)
	}

	info, err := readGistifyInfo()
	if err != nil {
		log.Fatalf("could not read gistinfo file %s: %v", gistifyFile, err)
	}

	var inputFiles []string
	for _, file := range files {
		if r.MatchString(file) {
			inputFiles = append(inputFiles, file)
		}
	}
	fmt.Printf("Files to be processed: %d\n", len(inputFiles))

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), tokenSource)
	client := github.NewClient(httpClient)

	currentUser, _, err := client.Users.Get(context.TODO(), "")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range inputFiles {
		f, err := os.Stat(file)
		if err != nil {
			continue
		}

		fcontent, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal("could not access file %s : %v", file, err)
		} else {
			fcontent := string(fcontent)
			if len(fcontent) == 0 {
				continue
			}

			filename := filepath.Base(file)
			newFile := true
			ginfo, exists := info[file]
			if exists {
				currMod := f.ModTime().Unix()
				if ginfo.Lastmod == currMod {
					fmt.Println("Skipping (file unchanged):", file)
					continue
				}
				newFile = false
			}
			gist := &github.Gist{}
			gist.Owner = currentUser
			gist.Description = github.String("")
			gist.Public = github.Bool(false)
			gist.Files = map[github.GistFilename]github.GistFile{}

			gist.Files[github.GistFilename(filename)] = github.GistFile{
				Content: github.String(fcontent),
			}

			if newFile {
				result, _, err := client.Gists.Create(context.TODO(), gist)
				if err != nil {
					log.Fatal(err)
				}
				info[file] = &gistinfo{ID: *result.ID, URL: *result.HTMLURL, Lastmod: f.ModTime().Unix(), Public: false, Filename: file}

				fmt.Println("Creating gist:", file+":", *result.HTMLURL)
			} else {
				_, _, err := client.Gists.Edit(context.TODO(), ginfo.ID, gist)
				if err != nil {
					log.Fatal(err)
				}
				info[file] = &gistinfo{ID: ginfo.ID, URL: ginfo.URL, Lastmod: f.ModTime().Unix(), Public: false, Filename: file}

				fmt.Println("Updating gist:", file+":", ginfo.URL)
			}

			if err != nil {
				log.Fatal(err)
			}
		}
	}

	_, err = saveGistifyInfo(info)
	if err != nil {
		log.Fatalf("could not save gistify file %s: %v", gistifyFile, err)
	}
}

func saveGistifyInfo(info map[string]*gistinfo) (bool, error) {
	f, err := os.Create(gistifyFile)
	if err != nil {
		return false, err
	}
	defer f.Close()

	wr := bufio.NewWriter(f)
	defer wr.Flush()

	enc := json.NewEncoder(wr)
	enc.SetIndent("", "    ")

	err = enc.Encode(info)
	if err != nil {
		return false, err
	}
	return true, err
}

func readGistifyInfo() (map[string]*gistinfo, error) {
	var info map[string]*gistinfo
	info = make(map[string]*gistinfo)

	if _, err := os.Stat(gistifyFile); os.IsNotExist(err) {
		return info, nil
	}

	f, err := os.Open(gistifyFile)
	if err != nil {
		return info, err
	}
	defer f.Close()

	err = json.NewDecoder(bufio.NewReader(f)).Decode(&info)

	return info, err
}
