package module

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	siteKey string = "6Lef5iQTAAAAAKeIvIY-DeexoO3gj7ryl9rLMEnn"
	//registerAPI string = "https://discordapp.com/register"
)

// Captcha solver
func Captcha(apiKey string) string {
	var reCaptcha string
	capurl := "http://2captcha.com/in.php?key=" + apiKey + "&method=userrecaptcha&googlekey=" + siteKey + "&pageurl=" + registerAPI

	req, err := http.Post(capurl, "application/json", nil)
	resp, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	data := string(resp)
	captcha := ""
	if strings.Contains(data, "|") == false {
		Log(data, "2CAPTCHA", "warn")
		return ""
	}

	Log("Waiting For reCAPTCHA v3 key.", "Log", "log")
	captcha = strings.Split(data, "|")[1]

	Resurl := "http://2captcha.com/res.php?key=" + apiKey + "&action=get&id=" + captcha
	captchaReady := false
	for captchaReady == false {
		response, err := http.Get(Resurl)
		if err != nil {
			fmt.Println(err)
		}
		recaptchaAnswer, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
		}
		if strings.Contains(string(recaptchaAnswer), "CAPCHA_NOT_READY") {
			captchaReady = false
		} else {
			captchaReady = true
			reCaptcha = strings.Split(string(recaptchaAnswer), "|")[1]
			return reCaptcha
		}
		time.Sleep(3 * time.Second)
	}
	return ""
}
