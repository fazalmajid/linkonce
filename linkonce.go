package main

import (
	"bufio"
	"bytes"
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

var (
	verbose   *bool
	dest      *string
	stateFile *string
	already   map[string]bool
)

func walk(path string, d fs.DirEntry, err error) error {
	if !d.IsDir() && !already[path] {
		if *verbose {
			log.Println(path)
		}
		newpath := filepath.Join(*dest, path)
		err := os.Link(path, newpath)
		if err != nil {
			destdir := filepath.Dir(newpath)
			err = os.MkdirAll(destdir, 0755)
			if err != nil {
				log.Fatal("could not create hard link dest dir", destdir,
					":", err)
			}
			err = os.Link(path, newpath)
			if err != nil {
				log.Fatal("could not create hard link", newpath,
					":", err)
			}
		}
		already[path] = true
	}
	return nil
}

func makeDest() {
	if *dest == "" {
		log.Fatal("must specify a destination directory")
	}
	err := os.MkdirAll(*dest, 0755)
	if err != nil {
		log.Fatal("error creating destination directory:", err)
	}
}

func readState() {
	f, err := os.Open(*stateFile)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(
		// based on bufio.ScanLines, just looking for NUL separators
		func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if i := bytes.IndexByte(data, 0); i >= 0 {
				// We have a full NULL-separated path.
				return i + 1, data[0:i], nil
			}
			// If we're at EOF, we have a final, non-terminated path.
			// Return it.
			if atEOF {
				return len(data), data, nil
			}
			// Request more data.
			return 0, nil, nil
		},
	)
	for scanner.Scan() {
		already[scanner.Text()] = true
	}
}

func saveState() {
	f, err := os.OpenFile(*stateFile+".tmp",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("could not create temp state file:", err)
	}
	buf := bufio.NewWriter(f)
	for dir, _ := range already {
		_, err = buf.WriteString(dir)
		if err != nil {
			log.Fatal("error writing state file:", err)
		}
		err = buf.WriteByte(0)
		if err != nil {
			log.Fatal("error writing state file:", err)
		}
	}
	err = buf.Flush()
	if err != nil {
		log.Fatal("error flushing state file:", err)
	}
	err = f.Close()
	if err != nil {
		log.Fatal("error closing state file:", err)
	}
	err = os.Rename(*stateFile+".tmp", *stateFile)
	if err != nil {
		log.Fatal("error renaming temp state file:", err)
	}
}

func main() {
	verbose = flag.Bool("v", false, "verbose output")
	dest = flag.String("d", "", "destination directory")
	stateFile = flag.String("s", ".linkonce",
		"state file to remember already linked files")
	flag.Parse()

	makeDest()

	already = make(map[string]bool, 20000)
	readState()

	err := filepath.WalkDir(".", walk)
	if err != nil {
		log.Fatal("error walking tree:", err)
	}

	saveState()
}
