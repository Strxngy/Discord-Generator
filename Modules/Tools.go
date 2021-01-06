package module

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Colors
var (
	blue    = color.New(color.FgBlue).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	orange  = color.New(color.FgHiRed).SprintFunc()
	yellow  = color.New(color.FgHiYellow).SprintFunc()
	hiblue  = color.New(color.FgHiBlue).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
)

const (
	letterBytes   = "abcdPQRSTUVefstuvwxyzABCDEFGHghijklmnopqrIJKLMNOWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	replacement   = ""
)

// vers
var (
	DeveloperModeSet bool
	ShowWarnsSet     bool
	ShowLogsSet      bool
	apiKey           string
	replacer         = strings.NewReplacer("\n", replacement)
	sreplacer        = strings.NewReplacer("\r", replacement)
)

// Revnewline removes /r and /n
func Revnewline(s string) string {
	s = replacer.Replace(s)
	return sreplacer.Replace(s)
}

// Password generator base64
func Password(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		log.Fatal(err)
	}
	v := base64.URLEncoding.EncodeToString(b)
	if err != nil {
		log.Fatal(err)
	}
	return v
}

// Save hits function
func Save(message string, filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		log.Fatal("We got error ", err)
	}

	defer file.Close()

	file.WriteString(message + "\n")
	file.Close()
}

// Configuration : ProxyFile string, Combolist string, TwoCaptcha string, Invite string, Speed:Threads int
type Configuration struct {
	ProxyFile     string `json:"ProxiesFile"`
	Combolist     string `json:"CombolistFile"`
	TwoCaptcha    string `json:"TwoCaptchaKey"`
	Invite        string `json:"InviteCode"`
	Speed         int    `json:"Threads"`
	ShowWarns     bool   `json:"ShowWarns"`
	DeveloperMode bool   `json:"DeveloperMode"`
	ShowLogs      bool   `json:"ShowLogs"`
}

// LoadConfiguration E.g: config.ProxyFile => path/name.txt
func LoadConfiguration(filename string) (Configuration, error) {
	var config Configuration
	configFile, err := os.Open(filename)
	defer configFile.Close()
	if err != nil {
		return config, err
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	ShowWarnsSet = config.DeveloperMode
	DeveloperModeSet = config.DeveloperMode
	ShowLogsSet = config.ShowLogs
	apiKey = config.TwoCaptcha
	return config, err
}

// Log print colorful chars in console
func Log(message string, box string, colorsidkthis string) {
	switch colorsidkthis {
	case "log":
		if ShowLogsSet == true {
			fmt.Fprintf(color.Output, "[%s] %s\n", blue("Log"), message)
		}
	case "warn":
		if ShowWarnsSet == true {
			fmt.Fprintf(color.Output, "[%s] %s\n", yellow("Warn"), message)
		}
	case "fatal":
		if DeveloperModeSet == true {
			fmt.Fprintf(color.Output, "[%s] %s\n", red(box), message)
		}
	case "bad":
		if DeveloperModeSet == true {
			fmt.Fprintf(color.Output, "[%s] %s\n", orange(box), message)
		}
	case "success":
		fmt.Fprintf(color.Output, "[%s] %s\n", green(box), message)
	case "error":
		if DeveloperModeSet == true {
			fmt.Fprintf(color.Output, "[%s] %s\n", magenta(box), message)
		}
	}

}

// Logo func prints acsii art
func Logo() {
	art := `
    ____  _                       _   _____           _ 
   |  _ \(_)___  ___ ___  _ __ __| | |_   _|__   ___ | |
   | | | | / __|/ __/ _ \| '__/ _' |   | |/ _ \ / _ \| |
   | |_| | \__ \ (_| (_) | | | (_| |   | | (_) | (_) | |
   |____/|_|___/\___\___/|_|  \__,_|   |_|\___/ \___/|_|
   `
	fmt.Fprintf(color.Output, "%s\n\t\t %s %s%s\n\n\n", blue(art), hiblue("Developed by"), blue("@"), orange("Strxngy"))
}
