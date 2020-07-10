package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

//define once to close the channel
var once sync.Once

//TripperResult used for cli output
type TripperResult struct {
	DsnLookup    string `json:"dnsLookUp"`
	TCPConnect   string `json:"tcpConnect"`
	TLSHandshake string `json:"tlsHandshake"`
	Ttfb         string `json:"ttfb"`
	Took         string `json:"took"`
}

//Tripper struct used to define a new type which will be
//unmarshaled later on in the application.
type Tripper struct {
	DNS struct {
		Start         string `json:"start"`
		timeStart     time.Time
		End           string `json:"end"`
		TotalTime     string `json:"totalTime"`
		timeTotalTime int64
		Host          string       `json:"host"`
		Address       []net.IPAddr `json:"address"`
		Error         error        `json:"error"`
	} `json:"dns"`
	Dial struct {
		Start         string `json:"start"`
		timeStart     time.Time
		timeEnd       time.Time
		End           string `json:"end"`
		TotalTime     string `json:"totalTime"`
		timeTotalTime int64
	} `json:"dial"`
	Connection struct {
		Time    string `json:"time"`
		timeCon int64
	} `json:"connection"`
	WroteAllRequestHeaders struct {
		Time string `json:"time"`
	} `json:"wrote_all_request_header"`
	WroteAllRequest struct {
		Time string `json:"time"`
	} `json:"wrote_all_request"`
	FirstReceivedResponseByte struct {
		TimeToFirstByteResponse string `json:"timeToFirstByte"`
		timeTotalTime           int64
		Time                    string `json:"time"`
	} `json:"first_received_response_byte"`
}

var (
	// Command line flags.
	wordPtr string
	boolPtr bool
	debug   bool
)

//silent log messages
func init() {
	//Parse flag parameters
	flag.StringVar(&wordPtr, "url", "https://www.google.com", "a url")
	flag.BoolVar(&boolPtr, "json", false, "if set, output a json result")
	flag.BoolVar(&debug, "debug", false, "set debug to true, and print out logs")

	flag.Usage = usage
	flag.Parse()

	if debug != true {
		log.SetOutput(ioutil.Discard)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL\n\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flag.PrintDefaults()
}

func main() {
	start := time.Now()
	//number of goroutines
	gr := 150
	c := make(chan *Tripper, gr)

	for i := 0; i < int(gr); i++ {
		go makeRequest(wordPtr, c)
	}

	i := 1
	var dnsAvgR, dialAvgR, tlsAvg, ttfb int64
	for v := range c {
		if i == int(gr) {
			once.Do(func() {
				close(c)
			})
		}

		dnsAvgR = dnsAvgR + v.DNS.timeTotalTime
		dialAvgR = dialAvgR + v.Dial.timeTotalTime
		tlsAvg = tlsAvg + v.Connection.timeCon
		ttfb = ttfb + v.FirstReceivedResponseByte.timeTotalTime
		i++
	}

	tJSON := TripperResult{}
	tJSON.DsnLookup = (time.Duration(dnsAvgR/int64(i)) * time.Nanosecond).String()
	tJSON.TCPConnect = (time.Duration(dialAvgR/int64(i)) * time.Nanosecond).String()
	tJSON.Ttfb = (time.Duration(ttfb/int64(i)) * time.Nanosecond).String()
	tJSON.TLSHandshake = (time.Duration(tlsAvg/int64(i)) * time.Nanosecond).String()
	tJSON.Took = (time.Since(start)).String()

	printResult(tJSON, boolPtr, wordPtr)
}

func printResult(tr TripperResult, jsonOutput bool, wordPtr string) {
	if !jsonOutput {
		//Print report.
		fmt.Println()
		fmt.Printf("%s \t %s\r\n", color.GreenString("Check connection data for : "), color.YellowString(wordPtr))
		fmt.Println()
		fmt.Printf("%s \t %s\r\n", "DNS lookup    ", tr.DsnLookup)
		fmt.Printf("%s \t %s\r\n", "TCP connection", tr.TCPConnect)
		fmt.Printf("%s \t %s\r\n", "TLS handshake ", tr.TLSHandshake)
		fmt.Printf("%s \t %s\r\n", "ttfb       ", tr.Ttfb)

		fmt.Printf("\r\nTripper took \t%s\r\n", color.MagentaString(tr.Took))
	} else {
		// Print json report.
		data, err := json.MarshalIndent(tr, "", "    ")
		if err != nil {
			log.Panic(err)
		}
		fmt.Println(string(data))
	}
}

func makeRequest(URL string, c chan *Tripper) {
	// Create trace struct.
	trace, tripper := trace()

	// Prepare request with trace attached to it.
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		log.Fatalln("request error", err)
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// MAke a request.
	res, err := client().Do(req)
	if err != nil {
		log.Fatalln("client error", err)
	}
	defer res.Body.Close()

	c <- tripper
}

func client() *http.Client {
	return &http.Client{
		Transport: transport(),
	}
}

func transport() *http.Transport {
	return &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   tlsConfig(),
	}
}

func tlsConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}

func trace() (*httptrace.ClientTrace, *Tripper) {
	d := &Tripper{}
	t := &httptrace.ClientTrace{

		DNSStart: func(info httptrace.DNSStartInfo) {
			d.DNS.timeStart = time.Now()
			log.Println(d.DNS.timeStart.UTC().String(), "dns start")
			d.DNS.Start = d.DNS.timeStart.UTC().String()
			d.DNS.Host = info.Host
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			t := time.Now()
			log.Println(t, "dns end")

			d.DNS.End = t.String()
			d.DNS.TotalTime = time.Since(d.DNS.timeStart).String()
			d.DNS.timeTotalTime = int64(time.Since(d.DNS.timeStart))
			d.DNS.Address = info.Addrs
			d.DNS.Error = info.Err
		},
		ConnectStart: func(network, addr string) {
			d.Dial.timeStart = time.Now()
			t := d.Dial.timeStart.UTC().String()
			log.Println(t, "dial start")
			d.Dial.Start = t
		},
		ConnectDone: func(network, addr string, err error) {
			t := time.Now()
			log.Println(t.UTC().String(), "dial end")
			d.Dial.End = t.UTC().String()
			d.Dial.timeEnd = t
			d.Dial.TotalTime = time.Since(d.Dial.timeStart).String()
			d.Dial.timeTotalTime = int64(time.Since(d.Dial.timeStart))

		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			t := time.Now()
			log.Println(t, "conn time")
			d.Connection.Time = t.UTC().String()
			d.Connection.timeCon = int64(t.Sub(d.Dial.timeEnd))
		},
		WroteHeaders: func() {
			t := time.Now().UTC().String()
			log.Println(t, "wrote all request headers")
			d.WroteAllRequestHeaders.Time = t
		},
		WroteRequest: func(wr httptrace.WroteRequestInfo) {
			t := time.Now().UTC().String()
			log.Println(t, "wrote all request")
			d.WroteAllRequest.Time = t
		},
		GotFirstResponseByte: func() {
			t := time.Now()
			log.Println(t.UTC().String(), "first received response byte")
			d.FirstReceivedResponseByte.TimeToFirstByteResponse = time.Since(d.Dial.timeStart).String()
			d.FirstReceivedResponseByte.timeTotalTime = int64(time.Since(d.Dial.timeStart))
			d.FirstReceivedResponseByte.Time = t.UTC().String()
		},
	}

	return t, d
}

func inputData(text string) string {
	var input string
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print(text)
	if scanner.Scan() {
		input = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Panic("Data not selected! Exiting....")
	}
	return input
}
