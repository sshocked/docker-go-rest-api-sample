package main

import (
	"bufio"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

type ProxyItem struct {
	proxyIp   string
	proxyPort string
}

type ProxyList []ProxyItem

func getLoadedProxy() ProxyList {
	//Загрузка .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

func getRotatedProxy(pList ProxyList, proxyOffset *int) ProxyItem {
	*proxyOffset++
	if *proxyOffset == len(pList) {
		*proxyOffset = 0
	}

	return pList[*proxyOffset]
}

func loadUrl(targetUrl string, pItem ProxyItem) *http.Response {
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		panic(err)
	}
	proxyString := "http://" + pItem.proxyIp + ":" + pItem.proxyPort
	client := getProxyAwareClient(proxyString)

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(contents))

	return res
}

func getProxyAwareClient(proxyString string) *http.Client {
	var dialer proxy.Dialer
	dialer = proxy.Direct
	//proxyServer, isSet := os.LookupEnv("HTTP_PROXY")
	proxyUrl, err := url.Parse(proxyString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid proxy url %q\n", proxyUrl)
	}
	dialer, err = proxy.FromURL(proxyUrl, proxy.Direct)

	// setup a http client
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial

	return httpClient
}

func main() {
	//Загружаем прокси из файла
	plist := getLoadedProxy()
	//Устанавливаем случайный индекс для начала ротации
	proxyOffset := getRandomProxyOffsetStart(len(plist))
	fmt.Println("Начальный индекс начала ротации:", proxyOffset, "ip:port", plist[proxyOffset].proxyIp, ":", plist[proxyOffset].proxyPort)
	loadUrl("https://tsum.ru/", getRotatedProxy(plist, &proxyOffset))

	//htmlData, err := ioutil.ReadAll(response.Body)

	//fmt.Println(string(htmlData))
}
