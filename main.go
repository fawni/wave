package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/logrusorgru/aurora/v3"
	"github.com/parnurzeal/gorequest"
)

type Config struct {
	CSRFToken string `json:"csrftoken,omitempty"`
	SessionID string `json:"sessionid,omitempty"`
	Output    string `json:"output,omitempty"`
	Proxies   string `json:"proxies,omitempty"`
}

type Username struct {
	Valid bool `json:"valid,omitempty"`
}

type Response struct {
	Username `json:"userName,omitempty"`
}

var cfg Config

func main() {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println(aurora.Red("failed to read config.json: " + err.Error()))
		os.Exit(1)
	}
	if err := json.Unmarshal(configFile, &cfg); err != nil {
		fmt.Println(aurora.Red("failed to unmarshal config.json: " + err.Error()))
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Output), 0755); err != nil {
		fmt.Println(aurora.Red("failed to create results directory: " + err.Error()))
		os.Exit(1)
	}
	res, err := os.Create(cfg.Output)
	if err != nil {
		fmt.Println(aurora.Red("failed to create results file: " + err.Error()))
		os.Exit(1)
	}
	names := read(os.Args[1])

	check(res, names)
}

func check(res *os.File, names []string) {
	for _, name := range names {
		_, err := strconv.Atoi(string(name[0]))
		if len(name) > 2 && len(name) <= 15 && err != nil {
			request(res, name)
		}
	}
}

func request(f *os.File, username string) {
	proxies := read(cfg.Proxies)
	r := gorequest.New().Proxy(proxies[rand.Intn(len(proxies))])
	url := "https://www.last.fm/join/partial/validate"
	cookie := fmt.Sprintf("csrftoken=%s; sessionid=%s", cfg.CSRFToken, cfg.SessionID)
	payload := fmt.Sprintf("userName=%s&csrfmiddlewaretoken=%s", username, cfg.CSRFToken)
	res, body, errs := r.Post(url).Set("referer", "https://www.last.fm/join").Set("cookie", cookie).Send(payload).End()
	if errs != nil {
		fmt.Println("http request failed: ", aurora.Red(errs))
	}
	if res.StatusCode == 403 {
		fmt.Println("csrftoken/sessionid invalid")
		os.Exit(1)
	}
	var data Response
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println(aurora.Red("failed to unmarshal response data: " + err.Error()))
	}

	switch data.Username.Valid {
	case false:
		fmt.Println(aurora.Red(username))
	case true:
		fmt.Printf("%s -> available!\n", aurora.Bold(aurora.Green(username)))
		f.WriteString(username + "\n")
	default:
		request(f, username)
	}
}

func read(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(aurora.Red("failed to open file: " + err.Error()))
		os.Exit(1)
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	var data []string
	for fileScanner.Scan() {
		data = append(data, fileScanner.Text())
	}
	return data
}
