package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"

	"github.com/atotto/clipboard"
	"github.com/howeyc/gopass"
)

type Note struct {
	Domain string `json:"domain"`
	Text   string `json:"text"`
}

type NoteCollection struct {
	Pool map[string]Note
}

func notes_file() string {
	// This won't work on Windows. But who cares about windows?
	return fmt.Sprintf("%s/.hap.json", os.Getenv("HOME"))
}

func read_notes() *NoteCollection {
	nc := new(NoteCollection)
	// Read from existing config file
	data, _ := ioutil.ReadFile(notes_file())
	json.Unmarshal(data, &nc.Pool)
	return nc
}

func display_notes(domain string) {
	nc := read_notes()
	keys := []string{}

	for key, _ := range nc.Pool {
		if nc.Pool[key].Domain == domain {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	for _, key := range keys {
		t, _ := time.Parse("20060102150405", key)
		fmt.Printf("%s => %s\n", t.Format("Jan 2, 2006 at 3:04 PM"), nc.Pool[key].Text)
	}
}

func add_note(domain string, text string) {
	nc := read_notes()

	// Create new note
	note := Note{}
	note.Domain = domain
	note.Text = text

	// Append new note to existing notes
	if nc.Pool == nil {
		nc.Pool = make(map[string]Note)
	}
	nc.Pool[time.Now().Format("20060102150405")] = note

	// Rewrite notes_file
	data, _ := json.Marshal(nc.Pool)
	ioutil.WriteFile(notes_file(), data, 0644)
}

func gen_pwd(data string, salt string) string {
	key := []byte(salt)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func main() {
	var domain, salt, master, note string = "", "", "", ""
	var pass_length int = 14

	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s domain [password-length]\n", os.Args[0])
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		// We do have domain
		domain = strings.TrimSpace(os.Args[1])
		display_notes(domain)
	}

	if len(os.Args) > 2 {
		// We do have password length
		p_length, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Password Length should be a number\n")
			os.Exit(1)
		}
		pass_length = p_length
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter salt. Hit enter to leave it blank: ")
	salt, _ = reader.ReadString('\n')
	salt = strings.TrimRight(salt, "\n")

	fmt.Printf("Enter note to save. Hit enter to leave it blank: ")
	note, _ = reader.ReadString('\n')
	note = strings.TrimRight(note, "\n")

	if len(note) > 0 {
		add_note(domain, note)
	}

	fmt.Printf("Enter Master password. Hit enter to abort: ")
	master = string(gopass.GetPasswdMasked()[:])

	if len(master) == 0 {
		os.Exit(1)
	}

	password := gen_pwd(domain+salt, master)[:pass_length]
	clipboard.WriteAll(password)
	fmt.Println(password, "copied to your clipboard")
}
