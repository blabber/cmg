// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/blabber/cmg/internal/backend"
	"github.com/blabber/cmg/lib"
)

var (
	port      = flag.String("port", "70", "The port cmg will be listening on")
	ip        = flag.String("ip", "", "The ip address cmg will be listening on (the listening socket will be bound to this ip; empty string for all)")
	host      = flag.String("host", "localhost", "The hostname or ip address cmg will be listening on (for use in menu items)")
	templates = flag.String("templates", "templates", "The directory containing the templates")
	recHttp   = flag.Bool("http", false, "link recordings via http, not gopher (saves bandwidth)")
)

const (
	categoryPrefix   = "/category/"
	conferencePrefix = "/conference/"
	eventPrefix      = "/event/"
	recordingPrefix  = "/recording/"
	urlPrefix        = "URL:"
)

const (
	typeDir    = "1"
	typeBinary = "9"
	typeHTML   = "h"
	typeInfo   = "i"
)

const cdn = "http://cdn.media.ccc.de/"

var (
	mainItem,
	raumZeitLaborItem,
	githubItem menuItem
)

var tmpl *template.Template

func main() {
	flag.Parse()

	tmpl = template.New("root").Funcs(template.FuncMap{"wrapDescription": wrapDescription})
	_, err := tmpl.ParseGlob(path.Join(*templates, "*.tmpl"))
	if err != nil {
		log.Panic(err)
	}

	l := lib.NewRateLimiter(time.Second/10, 100)
	defer l.Stop()

	ln, err := net.Listen("tcp", net.JoinHostPort(*ip, *port))
	if err != nil {
		log.Panic(err)
	}
	for {
		<-l.Throttle
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	s := bufio.NewScanner(conn)
	s.Scan()
	if s.Err() != nil {
		log.Print(s.Err())
		return
	}
	request := s.Text()
	log.Printf(`request: "%s"`, request)

	var err error
	var r io.Reader
	switch {
	case request == "" || request == "/":
		r, err = getMainMenu()
	case strings.HasPrefix(request, categoryPrefix):
		r, err = getCategoryMenu(strings.TrimPrefix(request, categoryPrefix))
	case strings.HasPrefix(request, urlPrefix):
		r, err = getUrlMenu(strings.TrimPrefix(request, urlPrefix))
	case strings.HasPrefix(request, conferencePrefix):
		r, err = getConferenceMenu(strings.TrimPrefix(request, conferencePrefix))
	case strings.HasPrefix(request, eventPrefix):
		r, err = getEventMenu(strings.TrimPrefix(request, eventPrefix))
	case strings.HasPrefix(request, recordingPrefix):
		r, err = getRecording(strings.TrimPrefix(request, recordingPrefix))
	default:
		err = fmt.Errorf(`no handler for request: "%s"`, request)
	}
	if closer, ok := r.(io.Closer); ok {
		defer closer.Close()
	}
	if err != nil {
		log.Print(err)
		r, err = getError(err)
		if err != nil {
			log.Panic(err)
		}
	}

	_, err = io.Copy(conn, r)
	if err != nil {
		log.Print(err)
		return
	}
}

type baseData struct {
	Host string
	Port string
}

type categoryData struct {
	baseData
	Slug  string
	Items []menuItem
}

type menuItem struct {
	Type     string
	Title    string
	Selector string
	Host     string
	Port     string
}

type conferenceData struct {
	baseData
	Slug     string
	Title    string
	MainItem menuItem
	Items    []menuItem
}

func (i menuItem) String() string {
	return fmt.Sprintf("%s%s\t%s\t%s\t%s", i.Type, i.Title, i.Selector, i.Host, i.Port)
}

type menuItemByTitle []menuItem

func (m menuItemByTitle) Len() int           { return len(m) }
func (m menuItemByTitle) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m menuItemByTitle) Less(i, j int) bool { return m[i].Title < m[j].Title }

func newMenuItem(t, title, selector string) menuItem {
	return menuItem{t, title, selector, *host, *port}
}

func ensureTrailingSlash(s string) string {
	if len(s) == 0 {
		return "/"
	}

	if s[len(s)-1:] == "/" {
		return s
	}

	return fmt.Sprintf("%s/", s)
}

func getCategoryData(category string) (*categoryData, error) {
	cs, err := backend.GetConferences()
	if err != nil {
		return nil, err
	}

	ic := make(chan menuItem)

	go func() {
		var wg sync.WaitGroup

		for _, c := range cs {
			cSlug := ensureTrailingSlash(c.Slug)
			categorySlash := ensureTrailingSlash(category)

			if strings.HasPrefix(cSlug, categorySlash) || category == "" {
				trailingString := strings.TrimPrefix(cSlug, categorySlash)
				trailingElements := strings.Split(trailingString, "/")
				nextElement := trailingElements[0]

				var checkSlug string
				if category == "" {
					checkSlug = nextElement
				} else {
					checkSlug = fmt.Sprintf("%s/%s", category, nextElement)
				}

				if c.Slug == checkSlug {
					wg.Add(1)
					go func(c backend.Conference) {
						defer wg.Done()

						id := c.Url[strings.LastIndex(c.Url, "/")+1:]

						bc, err := backend.GetConferenceById(id)
						if err != nil {
							log.Print(err)
						}

						if bc.Events != nil && len(bc.Events) > 0 {
							ic <- newMenuItem(typeDir, c.Title,
								fmt.Sprintf("%s%s", conferencePrefix, id))
						} else {
							log.Printf(`skipping empty conference "%s", %s`,
								c.Title, id)
						}
					}(c)
				} else {
					var selector string
					if category == "" {
						selector = fmt.Sprintf("%s%s", categoryPrefix, nextElement)
					} else {
						selector = fmt.Sprintf("%s%s/%s", categoryPrefix, category,
							nextElement)
					}
					ic <- newMenuItem(typeDir, nextElement, selector)
				}
			}
		}

		wg.Wait()
		close(ic)
	}()

	itemSet := make(map[menuItem]bool)
	for i := range ic {
		itemSet[i] = true
	}

	if len(itemSet) == 0 {
		return nil, fmt.Errorf(`unknown category: "%s"`, category)
	}

	var items []menuItem
	for i := range itemSet {
		items = append(items, i)
	}

	sort.Sort(menuItemByTitle(items))

	data := &categoryData{
		baseData{
			*host,
			*port,
		},
		category,
		items,
	}

	return data, nil
}

func getMainMenu() (io.Reader, error) {
	data, err := getCategoryData("")
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(b, "main.tmpl", data)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func getCategoryMenu(category string) (io.Reader, error) {
	data, err := getCategoryData(category)
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(b, "category.tmpl", data)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func getUrlMenu(url string) (io.Reader, error) {
	b := new(bytes.Buffer)
	err := tmpl.ExecuteTemplate(b, "url.tmpl", url)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func getConferenceData(id string) (*conferenceData, error) {
	c, err := backend.GetConferenceById(id)
	if err != nil {
		return nil, err
	}

	data := &conferenceData{
		baseData: baseData{
			*host,
			*port,
		},
		Slug:     c.Slug,
		Title:    c.Title,
		MainItem: mainItem,
	}

	for _, e := range c.Events {
		id := e.Url[strings.LastIndex(e.Url, "/")+1:]
		selector := fmt.Sprintf("%s%s", eventPrefix, id)
		data.Items = append(data.Items, newMenuItem(typeDir, e.Title, selector))
	}

	if len(data.Items) == 0 {
		return nil, fmt.Errorf(`unknown conference: "%s"`, id)
	}

	sort.Sort(menuItemByTitle(data.Items))

	return data, nil
}

func getConferenceMenu(id string) (io.Reader, error) {
	data, err := getConferenceData(id)
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(b, "conference.tmpl", data)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type eventData struct {
	baseData
	Title           string
	Subtitle        string
	Description     string
	Language        string
	Persons         string
	Tags            string
	Date            string
	Length          string
	RecordingsAudio []menuItem
	RecordingsVideo []menuItem
}

func getEventData(id string) (*eventData, error) {
	e, err := backend.GetEvent(id)
	if err != nil {
		return nil, err
	}

	data := &eventData{
		baseData: baseData{
			*host,
			*port,
		},
		Title:       e.Title,
		Subtitle:    e.Subtitle,
		Description: e.Description,
		Language:    e.Language,
		Persons:     strings.Join(e.Persons, ", "),
		Tags:        strings.Join(e.Tags, ", "),
		Date:        e.Date,
	}

	for _, r := range e.Recordings {
		var m menuItem
		if *recHttp {
			s := fmt.Sprintf("%s%s", urlPrefix, r.RecordingUrl)
			m = newMenuItem(typeHTML, r.Filename, s)
		} else {
			s := fmt.Sprintf("%s%s", recordingPrefix, strings.TrimPrefix(r.RecordingUrl, cdn))
			m = newMenuItem(typeBinary, r.Filename, s)
		}

		switch {
		case strings.HasPrefix(r.MimeType, "v"):
			data.RecordingsVideo = append(data.RecordingsVideo, m)
		case strings.HasPrefix(r.MimeType, "a"):
			data.RecordingsAudio = append(data.RecordingsAudio, m)
		}
	}

	if len(data.RecordingsVideo) == 0 && len(data.RecordingsAudio) == 0 {
		return nil, fmt.Errorf(`unknown event: "%s"`, id)
	}

	sort.Sort(menuItemByTitle(data.RecordingsVideo))
	sort.Sort(menuItemByTitle(data.RecordingsAudio))

	// Human readable length
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", e.Length))
	if err != nil {
		log.Print(err)
		data.Length = fmt.Sprintf("%d", duration)
	} else {
		data.Length = fmt.Sprintf("%v", duration)
	}

	return data, nil
}

func getEventMenu(id string) (io.Reader, error) {
	data, err := getEventData(id)
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(b, "event.tmpl", data)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func wrapDescription(d string) string {
	const width = 72
	const tail = "\tfake\tfake\t70\n"

	if d == "" {
		return d
	}

	d = strings.Replace(d, "\t", " ", -1)
	ss := strings.Split(d, "\n")

	var b bytes.Buffer
	for _, s := range ss {
		l := 0

		sc := bufio.NewScanner(strings.NewReader(s))
		sc.Split(bufio.ScanWords)
		b.WriteString(typeInfo)
		for sc.Scan() {
			if l > 0 {
				b.WriteString(" ")
				l = l + 1
			}

			if l+len(sc.Text()) > width {
				b.WriteString(tail)
				b.WriteString(typeInfo)
				l = 0
			}

			b.WriteString(sc.Text())
			l = l + len(sc.Text())
		}
		b.WriteString(tail)
	}

	return strings.Trim(b.String(), "\n")
}

func getRecording(path string) (io.Reader, error) {
	resp, err := http.Get(fmt.Sprintf("%s%s", cdn, path))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func getError(err error) (io.Reader, error) {
	data := struct {
		ErrorItem menuItem
		MainItem  menuItem
	}{
		newMenuItem(typeInfo, err.Error(), "fake"),
		mainItem,
	}

	b := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(b, "error.tmpl", data)
	if err != nil {
		return nil, err
	}

	return b, nil
}
