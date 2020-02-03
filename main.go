/*
Copyright 2019 github.com/ucirello

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Command alfredemoji downloads Unicode emoji files and generates an Alfre3
// compatible emoji snippet pack.
package main // import "cirello.io/alfredemoji"

import (
	"archive/zip"
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const unicodeVersion = "13.0"
const emojiSetURL = "https://unicode.org/Public/emoji/" + unicodeVersion + "/emoji-test.txt"

func main() {
	log.SetFlags(0)
	fd, err := os.Create("Emoji Pack (Unicode " + unicodeVersion + ").alfredsnippets")
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
	zipFD := zip.NewWriter(fd)
	defer zipFD.Close()
	resp, err := http.Get(emojiSetURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		l := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(l, "#") || l == "" || !strings.Contains(l, "; fully-qualified") {
			continue
		}
		parts := strings.SplitN(l[strings.Index(l, "# ")+2:], " ", 2)
		if len(parts) != 2 {
			log.Println("skipping", l)
		}
		emoji, name := parts[0], parts[1]
		s := newSnippet(emoji, name)
		if err := s.store(zipFD); err != nil {
			log.Println("cannot store in zip file (", s.Snippet.Name, "):", err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

type snippet struct {
	Snippet struct {
		Snippet string `json:"snippet"`
		UID     string `json:"uid"`
		Name    string `json:"name"`
		Keyword string `json:"keyword"`
	} `json:"alfredsnippet"`
}

func (s snippet) store(zipFD *zip.Writer) error {
	fn := fmt.Sprintf("%v [%v].json", s.Snippet.Name, s.Snippet.UID)
	w, err := zipFD.Create(fn)
	if err != nil {
		return fmt.Errorf("cannot create file in zip: %v", err)
	}
	if err := json.NewEncoder(w).Encode(s); err != nil {
		return fmt.Errorf("cannot write file in zip: %v", err)
	}
	if err := zipFD.Flush(); err != nil {
		return fmt.Errorf("cannot flush after file creation in zip: %v", err)
	}
	return nil
}

func newSnippet(emoji, name string) snippet {
	s := snippet{}
	s.Snippet.Snippet = emoji
	s.Snippet.UID = uuid()
	s.Snippet.Name = name
	normalizedName := strings.ToLower(name)
	normalizedName = strings.ReplaceAll(normalizedName, " ", "-")
	normalizedName = strings.ReplaceAll(normalizedName, ":", "")
	s.Snippet.Keyword = ":" + normalizedName + ":"
	return s
}

func uuid() string {
	b := make([]byte, 16)
	if n, err := rand.Read(b); err != nil || n != 16 {
		panic(err)
	}
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
