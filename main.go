package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"
)

const nitterUrl string = "https://nitter.cz/"

type Post struct {
	PubDate time.Time
	Image   Image
	Link    string
	Content template.HTML
}

type Posts struct {
	posts []Post
	sync.Mutex
}

func main() {
	feeds := make([]string, 0)

	accountsFile, err := os.Open("accounts.txt")
	if err != nil {
		fmt.Printf("Could not open accounts.txt: %v\n", err)
		os.Exit(1)
	}
	defer accountsFile.Close()

	scanner := bufio.NewScanner(accountsFile)
	for scanner.Scan() {
		feedUrl := fmt.Sprintf("%s%s/rss", nitterUrl, scanner.Text())
		feeds = append(feeds, feedUrl)
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from accounts.txt: %v\n", err)
		os.Exit(1)
	}

	p := Posts{}
	var wg sync.WaitGroup
	wg.Add(len(feeds))
	for _, feed := range feeds {
		go p.fetchPosts(feed, &wg)
	}

	wg.Wait()
	// Sort from latest to earliest.
	slices.SortFunc(p.posts, func(a Post, b Post) int {
		if a.PubDate.Before(b.PubDate) {
			return 1
		}
		return -1
	})

	outputFile, err := os.Create("output.html")
	if err != nil {
		fmt.Printf("Could not create output file: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	t, err := template.ParseFiles("template.html")
	if err != nil {
		fmt.Printf("Failed to create HTML template: %v\n", err)
		os.Exit(1)
	}
	err = t.Execute(outputFile, p.posts)
	if err != nil {
		fmt.Printf("Could not write to output file: %v\n", err)
		os.Exit(1)
	}
}

type Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Creator     string   `xml:"dc:creator"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate"`
	Guid        string   `xml:"guid"`
	Link        string   `xml:"link"`
}

type Image struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title"`
	Link    string   `xml:"link"`
	Url     string   `xml:"url"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Image   Image    `xml:"image"`
	Items   []Item   `xml:"item"`
}

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

func (p *Posts) fetchPosts(feedUrl string, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(feedUrl)
	if err != nil {
		fmt.Printf("Encountered an error while fetching feed %s: %v\n", feedUrl, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Non-200 status code: %d for feed %s\n", resp.StatusCode, feedUrl)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Encountered an error while reading feed %s: %v\n", feedUrl, err)
		return
	}

	var r Rss
	err = xml.Unmarshal(body, &r)
	if err != nil {
		fmt.Printf("Encountered an error while parsing XML for feed %s: %v\n", feedUrl, err)
		return
	}

	xmlPosts := make([]Post, 0, len(r.Channel.Items))
	for _, xmlPost := range r.Channel.Items {
		t, err := time.Parse(time.RFC1123, xmlPost.PubDate)
		if err != nil {
			fmt.Printf("Failed to parse time: %s for post %v. Skipping post.\n", xmlPost.PubDate, xmlPost)
			continue
		}

		xmlPosts = append(xmlPosts, Post{
			PubDate: t,
			Image:   r.Channel.Image,
			Link:    xmlPost.Link,
			Content: template.HTML(html.UnescapeString(xmlPost.Description)),
		})
	}

	p.Lock()
	p.posts = append(p.posts, xmlPosts...)
	p.Unlock()
}
