package main

import (
	"bufio"
	"database"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ProxyItem struct {
	proxyIp   string
	proxyPort string
}

type ProxyList []ProxyItem

var pList ProxyList

var proxyOffset int

var concurrentLimit int

var concurrentActive = 0

var maxDepth int

var protocolWithHost string

var targetHost string

var itemRule string

func getLoadedProxy() ProxyList {
	//Чтение имени файла со списком из конфига
	proxyListFileName := os.Getenv("PROXY_LIST_FILE_NAME")
	//Считывание файла
	file, err := os.Open(proxyListFileName)

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	//По строчное сканирование файла
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string
	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	file.Close()

	var pList ProxyList
	for _, eachline := range txtlines {
		//В каждой строке парсим ip и порт и записываем в структуру ProxyItem -> ProxyList
		r, _ := regexp.Compile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(\d{4,})$`)
		for _, match := range r.FindAllStringSubmatch(eachline, 2) {
			pList = append(pList, ProxyItem{match[1], match[2]})
		}
	}

	return pList
}

func getRandomProxyOffsetStart(pLen int) int {
	//Для того чтобы при тестовых запусках не палить первые ip
	//Будем каждый раз начнинать со случайного offset'a
	//Устанавливаем рандомный сид
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(pLen)
}

func getRotatedProxy(pList ProxyList) ProxyItem {
	proxyOffset++
	if proxyOffset == len(pList) {
		proxyOffset = 0
	}

	return pList[proxyOffset]
}

func getFullLink(urlObj *url.URL) string {
	return protocolWithHost + urlObj.Path
}

func checkIsExternalLink(urlObj *url.URL) bool {
	target, _ := url.Parse(protocolWithHost)

	return urlObj.Host != "" && urlObj.Host != target.Host
}

func checkIsSamePageLink(urlObj *url.URL) bool {
	return urlObj.Path == "/" || urlObj.Host == "" && urlObj.Fragment != ""
}

func checkIsTelLink(urlObj *url.URL) bool {
	return urlObj.Scheme == "tel"
}

func checkIsItemLink(urlObj *url.URL) bool {
	r, _ := regexp.Compile(itemRule)

	return r.MatchString(urlObj.Path)
}

func checkIsLinkAlreadyScanned(urlObj *url.URL) bool {
	var middleWare database.Middleware
	database.DB.Where("Path = ?", urlObj.Path).First(&middleWare)
	var item database.Item
	database.DB.Where("Path = ?", urlObj.Path).First(&item)

	return middleWare.ID != 0 || item.ID != 0
}

func loadUrl(targetUrl string, linkTitle string, depth int) {
	fmt.Println("Parsing page", linkTitle)
	urlObj, err := url.Parse(targetUrl)
	if err != nil {
		panic(err)
	}
	if checkIsLinkAlreadyScanned(urlObj) {
		fmt.Println("Already scanned or in progress", targetUrl, "skipping")
		return
	}
	//We have to just stop on catch any external links
	if checkIsExternalLink(urlObj) {
		fmt.Println("Get external link ", targetUrl, "skipping it")
		return
	}

	if checkIsTelLink(urlObj) {
		fmt.Println("Get telephone link ", targetUrl, "skipping it")
		return
	}

	if checkIsSamePageLink(urlObj) {
		fmt.Println("Get same page link ", targetUrl, "skipping it")
		return
	}

	fullLink := getFullLink(urlObj)
	req, err := http.NewRequest("GET", fullLink, nil)
	if err != nil {
		panic(err)
	}
	//pItem := getRotatedProxy(pList)
	//proxyString := "http://" + pItem.proxyIp + ":" + pItem.proxyPort

	//proxyURL, err := url.Parse(proxyString)
	checkError("parse proxy url", err)

	//transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	client := &http.Client{}
	//client := &http.Client{Transport: transport}

	checkError("", err)

	res, err := client.Do(req)
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if checkIsItemLink(urlObj) {
		createItem(urlObj.Path, linkTitle, string(contents))

		//While reach an item, we have to stop scanning
		return
	}
	fmt.Println("Found middleware link:", urlObj.Path, linkTitle)
	// Save to db middleware for able continue the progress if progress was broken
	createMiddleWare(urlObj.Path, linkTitle)

	if depth > maxDepth {
		return
	}
	time.Sleep(time.Second * 5)

	scanMiddleware(urlObj, string(contents), depth)
}

func createItem(path string, linkTitle string, content string) {
	fmt.Println("Found item link:", path, linkTitle)
	// Item found, saving it
	item := database.Item{Title: linkTitle, Path: path, Body: content}
	database.DB.Create(&item)
}

func createMiddleWare(path string, linkTitle string) {
	middleWare := database.Middleware{Title: linkTitle, Path: path, Status: 0}
	database.DB.Create(&middleWare)
}

func scanMiddleware(urlObj *url.URL, rawHtml string, depth int) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(rawHtml))
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exist := s.Attr("href")
		linkTitle := s.Text()
		if exist {
			loadUrl(link, linkTitle, depth+1)
		}
	})
	// Marking middleware is finished for skipping while resume
	var middleWare database.Middleware
	database.DB.Where("Path = ? and Status = 1", urlObj.Path).First(&middleWare)
	if middleWare.ID != 0 {
		middleWare.Status = 0
		database.DB.Save(&middleWare)
	}
}

func preLoad() {
	protocol := os.Getenv("TARGET_PROTOCOL")
	host := os.Getenv("TARGET_HOST")
	// One time prepare protocol with host for all future requests
	protocolWithHost = protocol + "://" + host
	// LoadProxyList
	pList = getLoadedProxy()
	// Load random proxy offset for starting each time with random proxy
	// but keeping rotating one by one
	proxyOffset = getRandomProxyOffsetStart(len(pList))
	itemRule = os.Getenv("ITEM_PATTERN")
	targetHost = os.Getenv("TARGET_HOST")
	maxDepth, _ = strconv.Atoi(os.Getenv("MAX_DEPTH"))
	concurrentLimit, _ = strconv.Atoi(os.Getenv("CONCURRENT_LIMIT"))
}

func main() {

	// initialization db package
	_, err := database.Init()
	if err != nil {
		log.Println("connection to DB failed, aborting...")
		log.Fatal(err)
	}
	//Загрузка .env
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	preLoad()
	fmt.Println("Default index offset for proxy:", proxyOffset)
	depth := 1

	loadUrl(protocolWithHost+"/", "main page", depth)
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}
