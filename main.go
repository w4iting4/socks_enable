package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

func main() {
	// 定义命令行参数
	urlFlag := flag.String("u", "", "URL to fetch")
	proxyFileFlag := flag.String("p", "proxy.txt", "File containing proxy details")
	outputBodyFlag := flag.Bool("b", false, "Whether to output the response body to output.txt")
	outputFileFlag := flag.String("o", "output.txt", "Path to the output file")
	maxGoroutinesFlag := flag.Int("t", 1, "Maximum number of goroutines")
	flag.Parse()

	// 检查 URL 是否提供
	if *urlFlag == "" {
		fmt.Println("请使用 -u 标志提供一个 URL")
		return
	}

	// 从文件中读取代理配置
	file, err := os.Open(*proxyFileFlag)
	if err != nil {
		fmt.Printf("打开代理文件失败: %s\n", err)
		return
	}
	defer file.Close()

	// 打开或创建用于写入结果的文件
	outputFile, err := os.OpenFile(*outputFileFlag, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("打开输出文件失败: %s\n", err)
		return
	}
	defer outputFile.Close()

	var wg sync.WaitGroup
	var mutex sync.Mutex

	// 控制协程数量的信号量
	goroutineSem := make(chan struct{}, *maxGoroutinesFlag)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxyURL := scanner.Text()

		// 等待一个信号量槽位
		goroutineSem <- struct{}{}
		wg.Add(1)

		go func(purl string) {
			defer wg.Done()
			defer func() { <-goroutineSem }() // 释放信号量槽位

			_, user, pass, host, err := parseProxyURL(purl)
			if err != nil {
				fmt.Printf("解析代理 URL 失败: %s\n", err)
				return
			}

			// 配置 SOCKS 代理
			dialer, err := proxy.SOCKS5("tcp", host, &proxy.Auth{
				User:     user,
				Password: pass,
			}, proxy.Direct)
			if err != nil {
				fmt.Printf("配置代理失败: %s\n", err)
				return
			}

			transport := &http.Transport{
				Dial: dialer.Dial,
			}

			client := &http.Client{
				Transport: transport,
				Timeout:   5 * time.Second,
			}

			response, err := client.Get(*urlFlag)

			mutex.Lock()
			if err != nil {
				fmt.Printf("请求失败: %s\n", err)
				return
			}
			defer response.Body.Close()

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("读取响应体失败: %s\n", err)
				return
			}
			if *outputBodyFlag {
				fmt.Printf("'%s': status_code: %d response_body: %s\n", purl, response.StatusCode, string(body))
				_, err = outputFile.WriteString(fmt.Sprintf("'%s': status_code: %d response_body: %s\n", purl, response.StatusCode, string(body)))
			} else {
				fmt.Printf("'%s': status_code: %d\n", purl, response.StatusCode)
				_, err = outputFile.WriteString(fmt.Sprintf("'%s': status_code: %d\n", purl, response.StatusCode))
			}
			mutex.Unlock()

			if err != nil {
				fmt.Printf("写入文件时出错: %s\n", err)
			}
		}(proxyURL)
	}

	wg.Wait() // 等待所有协程完成

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取代理文件时出错: %s\n", err)
	}
}

func parseProxyURL(proxyURL string) (protocol, user, pass, host string, err error) {
	if strings.HasPrefix(proxyURL, "socks5://") {
		protocol = "socks5"
	} else if strings.HasPrefix(proxyURL, "socks4://") {
		protocol = "socks4"
	} else if strings.HasPrefix(proxyURL, "socks://") {
		protocol = "socks"
	} else {
		return "", "", "", "", fmt.Errorf("非法的代理格式")
	}

	proxyURL = strings.TrimPrefix(proxyURL, protocol+"://")
	atSplit := strings.Split(proxyURL, "@")
	if len(atSplit) != 2 {
		return "", "", "", "", fmt.Errorf("代理格式不正确")
	}

	userPass := atSplit[0]
	host = atSplit[1]

	upSplit := strings.Split(userPass, ":")
	if len(upSplit) != 2 {
		return "", "", "", "", fmt.Errorf("代理格式不正确")
	}
	user = upSplit[0]
	pass = upSplit[1]

	return protocol, user, pass, host, nil
}
