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

	"github.com/rjeczalik/wayback"
)

const (
	greenColor  string = "\033[32m"
	yellowColor string = "\033[33m"
	redColor    string = "\033[31m"
	resetColor  string = "\033[0m"
	layoutISO          = "2006-01-02"
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
		Timeout:   time.Second * 7}
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

func treatEndpoint(urlCheck string, entry string, endpoints []string, mode uint) []string {
	aux := strings.Split(entry, "Disallow:")
	if len(aux) <= 1 {
		return endpoints
	}
	endpoint := strings.Trim(aux[1], " ")
	if endpoint == "/" || endpoint == "*" || endpoint == "" {
		return endpoints
	}
	finalEndpoint := strings.Replace(endpoint, "*", "", -1)

	var finalPrint string
	for strings.HasPrefix(finalEndpoint, "/") {
		if len(finalEndpoint) >= 1 {
			finalEndpoint = finalEndpoint[1:] // Ex. /*/test or /*/*/demo
		} else {
			return endpoints
		}
	}
	for strings.HasSuffix(finalEndpoint, "/") {
		if len(finalEndpoint) >= 1 {
			finalEndpoint = finalEndpoint[0 : len(finalEndpoint)-1]
		} else {
			return endpoints
		}
	}
	if mode == 0 {
		finalPrint = urlCheck + "/" + finalEndpoint
	} else {
		finalPrint = finalEndpoint
	}

	if len(finalPrint) > 0 {
		if containsElement(endpoints, finalPrint) { // Avoid duplicates. Ex. view/ view/*
			return endpoints
		}
		endpoints = append(endpoints, finalPrint)
		fmt.Println(finalPrint)
	}
	return endpoints
}

func waybackMachine(urlCheck string, endpoints []string, verbose bool, mode uint) {
	currentYear := time.Now().Year()
	robots := "/robots.txt"
	startYear := currentYear - 5 // Check last 5 years (It ignores current year)
	url := "https://web.archive.org/web/"
	lastURL := ""
	for startYear < currentYear {
		timestamp, err := wayback.ParseTimestamp(layoutISO, fmt.Sprintf("%04d-01-01", startYear))
		wbMsg := fmt.Sprintf("%s. Wayback Machine Year %d", urlCheck, startYear)
		startYear += 1
		if err != nil {
			printError(fmt.Sprintf("WB %d", startYear), err.Error(), verbose)
			continue
		}
		_, t, err := wayback.AvailableAt(urlCheck, timestamp)
		if err != nil {
			printError(fmt.Sprintf("WB %d", startYear), err.Error(), verbose)
			continue
		}
		date := fmt.Sprintf("%04d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
		finalurl := url + date + "if_/" + urlCheck
		if finalurl == lastURL {
			printInfo("i", fmt.Sprintf("Skiping year %d. Same snapshot as previous", startYear-1), verbose)
			continue
		}
		lastURL = finalurl
		response, status := requestURL(finalurl+robots, verbose)

		if status == 200 {
			endpoints = parseResponse(wbMsg, urlCheck, response, endpoints, verbose, mode)
		} else {
			if status > 0 {
				printError(strconv.Itoa(status), wbMsg, verbose)
			}
		}
	}
}

func work(urlCheck string, mode uint, verbose bool, wayback bool) bool {
	var endpoints []string
	success := true
	robots := "/robots.txt"
	if strings.HasSuffix(urlCheck, "/") {
		urlCheck = urlCheck[0 : len(urlCheck)-1]
	}
	response, status := requestURL(urlCheck+robots, verbose)
	stringStatus := strconv.Itoa(status)
	if status == 200 {
		endpoints = parseResponse(urlCheck, urlCheck, response, endpoints, verbose, mode)
	} else {
		if status > 0 {
			printError(stringStatus, urlCheck, verbose)
		}
		success = false
	}
	if wayback {
		waybackMachine(urlCheck, endpoints, verbose, mode)
	}
	return success
}

func parseResponse(msg string, urlCheck string, response string, endpoints []string, verbose bool, mode uint) []string {
	printOk("200", msg, verbose)
	allDisallows := getDisallows(response)
	if len(allDisallows) == 0 {
		printInfo("i", "Nothing found here...", verbose)
		return endpoints
	}
	printInfo("i", fmt.Sprintf("Total entries marked as disallow: %d. Parsing and cleaning...", len(allDisallows)), verbose)
	for _, entry := range allDisallows {
		endpoints = treatEndpoint(urlCheck, entry[0], endpoints, mode)
	}
	return endpoints
}

func start(urlCheck string, mode uint, verbose bool, wayback bool) {
	if len(strings.Split(urlCheck, ".")) <= 1 {
		printError("-", "URL format error "+urlCheck, verbose)
		return
	}
	if !strings.HasPrefix(urlCheck, "http") {
		if !(work("https://"+urlCheck, mode, verbose, wayback)) {
			work("http://"+urlCheck, mode, verbose, wayback)
		}
	} else {
		work(urlCheck, mode, verbose, wayback)
	}
}

func main() {
	// Thanks to @remonsec for this idea
	// Check the tweet https://twitter.com/remonsec/status/1410481151433576449
	url := flag.String("u", "", "URL to extract endpoints marked as disallow in robots.txt file")
	mode := flag.Uint("m", 1, "Extract URLs (0) // Extract endpoints to generate a wordlist  (>1)")
	wayback := flag.Bool("wb", false, "Check Wayback Machine. Check 5 years (Slow mode)")
	verbose := flag.Bool("v", false, "Verbose mode. Displays additional information at each step")
	silent := flag.Bool("s", false, "Silen mode doesn't show banner")
	flag.Parse()
	banner(*silent)
	if *url != "" {
		start(*url, *mode, *verbose, *wayback)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			start(scanner.Text(), *mode, *verbose, *wayback)
			if *mode == 0 && *verbose {
				println("")
			}
		}
	}
	printOk("+", "Done", *verbose)
}
