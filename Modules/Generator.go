package module

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/corpix/uarand"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/go-resty/resty"
)

var (
	experimentsAPI string = "https://canary.discord.com/api/v6/experiments"
	registerAPI    string = "https://canary.discord.com/api/v6/auth/register"
	resendAPI      string = "https://canary.discord.com/api/v6/auth/verify/resend"
	verifyAPI      string = "https://canary.discord.com/api/v6/auth/verify"
	osinfo                = regexp.MustCompile(`(?:compatible;[^;]+; |Linux; )?(?P<os>[a-zA-Z]+(?: NT)?) ?(?P<os_ver>[\d\.]+)`)
	browserinfo           = regexp.MustCompile(`(?:KHTML, [\w ]+\)|compatible;|Android [\d\.]+; [\w \.;:]+\)) +(?P<browser>[\w ]+)[\/ ]+(?P<browser_ver>[\w\.+]+)`)
	c              *client.Client
)

func activation() string {
	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Println(err)
		return "nil"
	}
	// Get the last message
	if mbox.Messages == 0 {
		log.Println("No message in mailbox")
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(mbox.Messages, mbox.Messages)

	// Get the whole message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	msg := <-messages
	r := msg.GetBody(section)
	if r == nil {
		log.Fatal("Server didn't returned message body")
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Fatal(err)
	}

	rx := quotedprintable.NewReader(m.Body)

	body, err := ioutil.ReadAll(rx)
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile(`(?i)[(http(s)?):\/\/(www\.)?a-zA-Z0-9@:%._\\+~#=]{2,256}\.[a-z]{2,6}\b[-a-zA-Z0-9@:%_\+.~#?&//=]*`)
	matches := re.FindAllString(string(body), -1)
	if len(matches) != 0 {
		return matches[0]
	}
	return "<nil>"
}

func getexpr(client *resty.Client) (bool, string) {
	for i := 1; i < 3; i++ {
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			Get(experimentsAPI)

		if requesthandler(err) {
			return false, ""
		}

		body := resp.String()
		switch resp.StatusCode() {
		case 200:
			var result map[string]interface{}
			json.Unmarshal([]byte(string(body)), &result)
			return true, result["fingerprint"].(string)
		case 503:
			fmt.Println("[+] 503 Service Unavailable")
		case 500:
			fmt.Println("[+] 500 Unternal Server")
		case 403:
			fmt.Println("[+] 403 Forbidden")
		default:
			fmt.Println(body)
		}
		time.Sleep(3 * time.Second)
	}
	return false, "nil"
}

func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	return subMatchMap
}

func requesthandler(err error) bool {
	if err != nil {
		Log(fmt.Sprintf("%s", err), "Req", "error")
		return true
	}
	return false
}

// Creator Discord
func Creator(conn *client.Client, proxy *Reader, names *Reader, combo []string, invite string, apiKey string) {
	c = conn
	status := false
	var finger string
	client := resty.New()
	for status == false {
		useragent := uarand.GetRandom()
		proxyip := proxy.Rand()
		proxyStr := Revnewline(proxyip)
		if strings.HasPrefix(proxyStr, "http") == false {
			proxyStr = "http://" + proxyStr
		}
		client.SetProxy(proxyStr)
		client.Header.Add("User-Agent", useragent)
		status, finger = getexpr(client)
		if !status {
			continue
		} else {
			break
		}
	}
	username, email, password := names.Rand(), combo[0], Password(16)

	// Super := fmt.Sprintf(`{"os":"%s","browser":"%s","device":"","browser_user_agent":"%s","browser_version":"%s","os_version":"%s","referrer":"","referring_domain":"","referrer_current":"","referring_domain_current":"","release_channel":"stable","client_build_number":35540,"client_event_source":null}`, o["os"], b["browser"], useragent, b["browser_ver"], o["os_ver"])
	// Superx := base64.StdEncoding.EncodeToString([]byte(Super))

	// client.Header.Set("X-Fingerprint", finger)
	// client.Header.Set("X-Super-Properties", Superx)
	// client.Header.Del("X-Fingerprint")
	// client.Header.Del("X-Super-Properties")

	payload := fmt.Sprintf("{\"username\": \"%s\",\"email\": \"%s\",\"password\": \"%s\",\"invite\":\"%s\",\"fingerprint\":\"%s\",\"consent\":\"true\",\"captcha_key\":\"%s\"}", username, email, password, invite, finger, Captcha(apiKey))

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(payload).
		Post(registerAPI)

	if requesthandler(err) {
		return
	}

	var token string
	var result map[string]interface{}

	switch content := resp.String(); {
	case strings.Contains(content, "message"):
		json.Unmarshal([]byte(content), &result)
		Log(result["message"].(string), "Discord", "bad")
		return
	case strings.Contains(content, "token"):
		json.Unmarshal([]byte(content), &result)
		token = result["token"].(string)
		Log(token, "TOKEN", "success")
		Save(string(fmt.Sprintf("%s:%s:%s:%s", token, username, email, password)), "Output/tokens.txt")
	default:
		Log(content, "400", "bad")
		return
	}

	// verification part
	status, finger = getexpr(client)
	if !status {
		return
	}

	// resend Request
	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetHeader("X-Fingerprint", finger).
		Post(resendAPI)

	if requesthandler(err) {
		return
	}

	time.Sleep(30 * time.Second)
	haha := activation()
	if len(haha) <= 10 {
		Log("Getting Verification mail failed!", "IMAP", "error")
		return
	}

	rr, _ := regexp.Compile(`https?://[a-zA-Z0-9]+\.[a-zA-Z0-9]+\.com/ls/click\?upn=([\w-]+)`)
	verfurl := rr.FindString(haha)

	// get verification token
	resp, err = client.R().
		Get(verfurl)

	if requesthandler(err) {
		return
	}

	verftoken := resp.RawResponse.Request.URL.String()

	if !strings.Contains(verftoken, "token=") {
		fmt.Println("Doesent containtshit")
	}

	payload = fmt.Sprintf("{\"token\": \"%s\",\"captcha_key\": \"%s\"}", strings.Split(verftoken, "token=")[1], Captcha(apiKey))

	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(payload).
		Post(verifyAPI)

	if requesthandler(err) {
		return
	}

	fmt.Println(resp.String())

	Log("Email has succesfully been verified", "Discord", "log")

	files, err := ioutil.ReadDir("./Avatars")
	if err != nil {
		log.Fatal(err)
	}

	n1, err := rand.Int(rand.Reader, big.NewInt(int64(len(files))))
	if err != nil {
		panic(err)
	}

	f, err := os.Open(`./Avatars/` + files[n1.Int64()+0].Name())
	if err != nil {
		fmt.Println(err)
	}

	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)

	code := base64.StdEncoding.EncodeToString(content)                             // Encode as base64.
	jsonPayload := fmt.Sprint("{\"avatar\":\"data:image/png;base64,", code, "\"}") // ... The base64 image can be used as a data URI in a browser.

	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "*/*").
		SetHeader("Authorization", token).
		SetBody(jsonPayload).
		Patch("https://canary.discord.com/api/v6/users/@me")

	if requesthandler(err) {
		return
	}
	if resp.StatusCode() == 200 {
		Log("Avatar changed to "+files[n1.Int64()+0].Name(), "Discord", "log")
	} else {
		fmt.Println(resp.String())
		fmt.Println(resp.StatusCode())
	}

	return

}
