# WebCrawler
Web crawler written in Go.

 1. Get project dependencies:
 ```
    go get ./...
```
 2. Build the thing:
 ```
    go build webCrawler
 ```
 3. Run on a website:
 ```
    go run webCrawler "http://www.example.com"
 ```

Logs are always printed to stdout, while the output website map can be
redirected. This can be used to generate a file with the website map:
```
    go run webCrawler "http://www.example.com" > sitemap.example.com.txt
```