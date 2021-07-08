package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	greenColor  string = "\033[32m"
	yellowColor string = "\033[33m"
	redColor    string = "\033[31m"
	resetColor  string = "\033[0m"
)

func banner(silent bool) {
	if !silent {

		data := `
▄▄▄        ▄▄▄▄·       ▐▄• ▄ ▄▄▄▄▄▄▄▄   ▄▄▄·  ▄▄· ▄▄▄▄▄      ▄▄▄  
▀▄ █·▪     ▐█ ▀█▪▪      █▌█▌▪•██  ▀▄ █·▐█ ▀█ ▐█ ▌▪•██  ▪     ▀▄ █·
 ▀▀▄  ▄█▀▄ ▐█▀▀█▄ ▄█▀▄  ·██·  ▐█.▪▐▀▀▄ ▄█▀▀█ ██ ▄▄ ▐█.▪ ▄█▀▄ ▐▀▀▄ 
▐█•█▌▐█▌.▐▌██▄▪▐█▐█▌.▐▌▪▐█·█▌ ▐█▌·▐█•█▌▐█ ▪▐▌▐███▌ ▐█▌·▐█▌.▐▌▐█•█▌
.▀  ▀ ▀█▄▀▪·▀▀▀▀  ▀█▄▀▪•▀▀ ▀▀ ▀▀▀ .▀  ▀ ▀  ▀ ·▀▀▀  ▀▀▀  ▀█▄▀▪.▀  ▀
	
  > By @JosueEncinar
  > Extract endpoints marked as disallow in robots.txt file										   
	  `
		println(data)
	}
}

func printError(init string, msg string, verbose bool) {
	if verbose {
		println(redColor + "[" + init + "] " + resetColor + msg)
	}
}

func printInfo(init string, msg string, verbose bool) {
	if verbose {
		println(yellowColor + "[" + init + "] " + resetColor + msg)
	}
}

func printOk(init string, msg string, verbose bool) {
	if verbose {
		println(greenColor + "[" + init + "] " + resetColor + msg)
	}
}

func containsElement(list []string, element string) bool {
	exist := false
	for _, e := range list {
		if e == element {
			exist = true
			break
		}
	}
	return exist
}

func requestURL(url string, verbose bool) (string, int) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 6}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		printError("-", err.Error(), verbose)
		return "", 0
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux i686; rv:89.0) Gecko/20100101 Firefox/89.0")
	res, err := client.Do(req)
	if err != nil {
		printError("-", err.Error(), verbose)
		return "", 0
	}
	if res.Body != nil {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", res.StatusCode
		}
		return string(body), res.StatusCode

	}
	return "", res.StatusCode
}

func getDisallows(data string) [][]string {
	pattern := regexp.MustCompile("Disallow:\\s?.+")
	return pattern.FindAllStringSubmatch(data, -1)
}

func start(urlCheck string, mode uint, verbose bool) {
	// printInfo("i", "Checking "+urlCheck, verbose)
	if len(strings.Split(urlCheck, ".")) <= 1 {
		printError("-", "URL format error "+urlCheck, verbose)
		return
	}
	if !strings.HasPrefix(urlCheck, "http") {
		if !(work("https://"+urlCheck, mode, verbose)) {
			work("http://"+urlCheck, mode, verbose)
		}
	} else {
		work(urlCheck, mode, verbose)
	}
}

func work(urlCheck string, mode uint, verbose bool) bool {
	var allDataPrint []string
	robots := "/robots.txt"
	if strings.HasSuffix(urlCheck, "/") {
		urlCheck = urlCheck[0 : len(urlCheck)-1]
	}
	response, status := requestURL(urlCheck+robots, verbose)
	stringStatus := strconv.Itoa(status)
	if status != 200 {
		if status > 0 {
			printError(stringStatus, urlCheck, verbose)
		}
		return false
	}

	printOk(stringStatus, urlCheck, verbose)
	allDisallows := getDisallows(response)
	if len(allDisallows) == 0 && verbose {
		printInfo("i", "Nothing found here...", verbose)
	}
	for _, entry := range allDisallows {
		aux := strings.Split(entry[0], "Disallow:")
		if len(aux) <= 1 {
			continue
		}
		endpoint := strings.Trim(aux[1], " ")
		if endpoint == "/" || endpoint == "*" || endpoint == "" {
			continue
		}
		finalEndpoint := strings.Replace(endpoint, "*", "", -1)
		if containsElement(allDataPrint, finalEndpoint) { // Avoid duplicates. Ex. view/ view/*
			continue
		}
		allDataPrint = append(allDataPrint, finalEndpoint)
		var finalPrint string
		for strings.HasPrefix(finalEndpoint, "/") {
			finalEndpoint = finalEndpoint[1:] // Ex. /*/test or /*/*/demo
		}
		if mode == 0 {
			finalPrint = urlCheck + "/" + finalEndpoint
		} else {
			finalPrint = finalEndpoint // remove first /
		}
		if len(finalPrint) > 0 {
			fmt.Println(finalPrint)
		}
	}
	return true
}

func main() {
	// Thanks to @remonsec for this idea
	// Check the tweet https://twitter.com/remonsec/status/1410481151433576449
	url := flag.String("u", "", "URL to extract endpoints marked as disallow in robots.txt file")
	mode := flag.Uint("m", 1, "Extract URLs (0) // Extract endpoints to generate a wordlist  (>1)")
	verbose := flag.Bool("v", false, "Verbose mode. Displays additional information at each step")
	silent := flag.Bool("s", false, "Silen mode doesn't show banner")
	flag.Parse()
	banner(*silent)
	if *url != "" {
		start(*url, *mode, *verbose)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			start(scanner.Text(), *mode, *verbose)
			if *mode == 0 && *verbose {
				println("")
			}
		}
	}
	printOk("+", "Done", *verbose)
}
