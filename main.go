package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type gistMetadata struct {
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

func getGithubToken() (string, error) {
	token := os.Getenv("GISTIFY_TOKEN")
	if len(token) == 0 {
		return "", errors.New("ERROR: could not find GISTIFY_TOKEN environment variable set")
	}

	return token, nil
}

func searchAndfilterFiles(searchDir, searchPattern string) ([]string, error) {
	var files []string
	err := filepath.Walk(searchDir, visit(&files))
	if err != nil {
		return nil, fmt.Errorf("ERROR: could not search files in dir %s: %v", searchDir, err)
	}

	srchPtrn := os.Args[1] //Example: ".*.go"
	r, err := regexp.Compile(srchPtrn)
	if err != nil {
		return nil, fmt.Errorf("could not compile given regex %s: %v", srchPtrn, err)
	}

	var inputFiles []string
	for _, file := range files {
		if r.MatchString(file) {
			inputFiles = append(inputFiles, file)
		}
	}

	return inputFiles, nil
}

func newGist(currentUser *github.User, decription string, public bool, filename, content string) *github.Gist {
	gist := &github.Gist{}
	gist.Owner = currentUser
	gist.Description = github.String("")
	gist.Public = github.Bool(false)
	gist.Files = map[github.GistFilename]github.GistFile{}

	gist.Files[github.GistFilename(filename)] = github.GistFile{
		Content: github.String(content),
	}

	return gist
}

func writeMetadata(metadataFile string, info map[string]*gistMetadata) (bool, error) {
	f, err := os.Create(metadataFile)
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

func readMetadata(metadataFile string) (map[string]*gistMetadata, error) {
	var info map[string]*gistMetadata
	info = make(map[string]*gistMetadata)

	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		return info, nil
	}

	f, err := os.Open(metadataFile)
	if err != nil {
		return info, err
	}
	defer f.Close()

	err = json.NewDecoder(bufio.NewReader(f)).Decode(&info)

	return info, err
}

func main() {
	const metadataFile = ".gistify"

	token, err := getGithubToken()
	if err != nil {
		log.Fatal(err)
	}

	filepaths, err := searchAndfilterFiles(".", os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Files to be processed: %d\n", len(filepaths))

	metadata, err := readMetadata(metadataFile)
	if err != nil {
		log.Fatal(err)
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), tokenSource)
	client := github.NewClient(httpClient)
	currUser, _, err := client.Users.Get(context.TODO(), "")
	if err != nil {
		log.Fatal(err)
	}

	for _, relpath := range filepaths {
		f, err := os.Stat(relpath)
		if err != nil {
			continue
		}

		content, err := ioutil.ReadFile(relpath)
		if err != nil {
			log.Fatal("could not access file %s : %v", relpath, err)
		} else {
			content := string(content)
			if len(content) == 0 {
				fmt.Println("Skipping (no content):", relpath)
				continue
			}

			filename := filepath.Base(relpath)
			newFile := true
			gistMeta, exists := metadata[relpath]
			if exists {
				currMod := f.ModTime().Unix()
				if gistMeta.Lastmod == currMod {
					fmt.Println("Skipping (file unchanged):", relpath)
					continue
				}
				newFile = false
			}
			gist := newGist(currUser, "", false, filename, content)

			if newFile {
				result, _, err := client.Gists.Create(context.TODO(), gist)
				if err != nil {
					log.Fatal(err)
				}
				metadata[relpath] = &gistMetadata{ID: *result.ID, URL: *result.HTMLURL, Lastmod: f.ModTime().Unix(), Public: false, Filename: relpath}
				fmt.Println("Creating gist:", relpath+":", *result.HTMLURL)
			} else {
				_, _, err := client.Gists.Edit(context.TODO(), gistMeta.ID, gist)
				if err != nil {
					log.Fatal(err)
				}
				metadata[relpath] = &gistMetadata{ID: gistMeta.ID, URL: gistMeta.URL, Lastmod: f.ModTime().Unix(), Public: false, Filename: relpath}
				fmt.Println("Updating gist:", relpath+":", gistMeta.URL)
			}

			if err != nil {
				log.Fatal(err)
			}
		}
	}

	_, err = writeMetadata(metadataFile, metadata)
	if err != nil {
		log.Fatalf("could not save gistify metadata file %s: %v", metadataFile, err)
	}
}
