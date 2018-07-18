package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/kumakichi/goamf"
)

const (
	URL = "http://aqi.zjemc.org.cn/aqi/messagebroker/amf"
)

func main() {
	//You can send mutiple requests per time, like this:
	//
	//reader, header, err := amf.NewRequest([]interface{}{
	//	[]interface{}{"getAreaDayReprotData", "GisCommonDataUtil"},
	//	[]interface{}{"getAreaRealTimeReportData", "GisCommonDataUtil"},
	//})

	//Single request
	reader, header, err := amf.NewRequest("getAreaDayReprotData", "GisCommonDataUtil")
	checkError(err)

	b := requestWithHeader(reader, header)

	body, err := amf.ParseRespBody(b)
	checkError(err)
	fmt.Printf("Body size: %d,%v\n", len(body), body)
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
