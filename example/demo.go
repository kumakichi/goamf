package main

import (
	"net/http"

	"github.com/kumakichi/goamf"

	"fmt"
	"io"
	"io/ioutil"
	"log"
)

const (
	URL = "http://aqi.zjemc.org.cn/aqi/messagebroker/amf"
)

func main() {
	reader, header, err := amf.NewRequest("getAreaRealTimeReportData", "GisCommonDataUtil")
	checkError(err)

	b := requestWithHeader(reader, header)

	body, err := amf.ParseRespBody(b)
	checkError(err)
	fmt.Printf("Body size: %d\n", len(body))
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func requestWithHeader(reader io.Reader, header http.Header) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("POST", URL, reader)
	checkError(err)

	req.Header = header
	resp, err := client.Do(req)
	checkError(err)

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	checkError(err)

	return b
}
