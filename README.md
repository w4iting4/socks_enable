# socks_enable 
## 使用说明
为了解决大量socks无法确认是否可用的情况下而产生的小工具。
```
waiting4@waiting4deMacBook-Pro socks_enable % go run main.go -h
  -b    Whether to output the response body to output.txt
  -o string
        Path to the output file (default "output.txt")
  -p string
        File containing proxy details (default "proxy.txt")
  -t int
        Maximum number of goroutines (default 1)
  -u string
        URL to fetch
```
`-b 确认是否输出响应体`

`-o 输出文件`

`-p 指定socks5代理的文件位置，一行一个目标`

`-u 指定访问的目标地址，比如https://www.google.com、https://myip.ipip.net`

`-t 协程数量，最好不要超过socks文件的最大值`

运行效果
![image](https://github.com/w4iting4/socks_enable/assets/41547947/26f194d0-53d6-492a-82f2-6b172921cb2c)
