package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

func newGistMetadata(id, url string, modTime int64, isPublic bool, relpath string) *gistMetadata {
	return &gistMetadata{
		ID:       id,
		URL:      url,
		Lastmod:  modTime,
		Public:   isPublic,
		Filename: relpath,
	}
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

	r, err := regexp.Compile(searchPattern)
	if err != nil {
		return nil, fmt.Errorf("could not compile given regex %s: %v", searchPattern, err)
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
	gist.Public = github.Bool(public)
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

type gistStatus int

const (
	gistCreated gistStatus = iota
	gistRecreated
	gistUpdated
	gistFailed
)

func (s gistStatus) String() string {
	switch s {
	case gistCreated:
		return "Created  "
	case gistRecreated:
		return "Recreated"
	case gistUpdated:
		return "Updated  "
	case gistFailed:
		return "Failed   "
	default:
		return "Status?  "
	}
}

func processGist(id string, gist *github.Gist, client *github.Client) (*github.Gist, gistStatus, error) {
	if len(id) == 0 {
		g, _, err := client.Gists.Create(context.TODO(), gist)
		if err != nil {
			return nil, gistFailed, err
		}

		return g, gistCreated, nil
	}

	g, r, err := client.Gists.Edit(context.TODO(), id, gist)
	if err != nil {
		if r.StatusCode != http.StatusNotFound {
			return nil, gistFailed, err
		}

		g, _, err := client.Gists.Create(context.TODO(), gist)
		if err != nil {
			return nil, gistFailed, err
		}
		return g, gistRecreated, nil
	}

	return g, gistUpdated, nil

}

func main() {
	isPublic := flag.Bool("public", false, "sets visibility to public")

	flag.Usage = func() {
		var CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		fmt.Fprintf(CommandLine.Output(), "Usage of %s: %s [options] <pattern>\n", os.Args[0], os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	pattern := flag.Args()
	if len(pattern) == 0 {
		flag.Usage()
		return
	}

	token, err := getGithubToken()
	if err != nil {
		log.Fatal(err)
	}

	filepaths, err := searchAndfilterFiles(".", pattern[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Files to be processed: %d\n", len(filepaths))

	const metadataFile = ".gistify"
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

		bcontent, err := ioutil.ReadFile(relpath)
		if err != nil {
			log.Fatal("could not access file %s : %v", relpath, err)
		}
		content := string(bcontent)
		if len(content) == 0 {
			fmt.Println("Skipping (no content):", relpath)
			continue
		}

		id := ""
		modTime := f.ModTime().Unix()
		gistMeta, exists := metadata[relpath]

		if exists {
			if gistMeta.Lastmod == modTime {
				fmt.Println("Skipping ", gistMeta.URL, relpath, "(file unchanged)")
				continue
			}
			id = gistMeta.ID
		}

		gist := newGist(currUser, "", *isPublic, filepath.Base(relpath), content)
		result, status, err := processGist(id, gist, client)
		fmt.Println("" + status.String() + " " + *result.HTMLURL + " for file " + relpath)
		if err != nil {
			log.Fatal(err)
		}

		metadata[relpath] = newGistMetadata(*result.ID, *result.HTMLURL, modTime, *isPublic, relpath)

	}

	_, err = writeMetadata(metadataFile, metadata)
	if err != nil {
		log.Fatalf("could not save gistify metadata file %s: %v", metadataFile, err)
	}
}
