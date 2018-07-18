package amf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// NewRequest returns an io.Reader and a http.Header which can be used by http.NewRequest
//
// A single request, arguments: operation, destnation, body, targetUri, responseUri, dsEndpoint ...
//
//   NewRequest("getAreaUsedToDayReport", "GisCommonDataUtil")
//   // or
//   NewRequest("find24HourAqiInfoByCode", "dataPubService", []interface{}{"411729", "01"})
//
// A mutiple request:
//
//   args := []interface{}{
//		[]interface{}{"getAreaDayReprotData", "GisCommonDataUtil"},
//		[]interface{}{"getAreaRealTimeReportData", "GisCommonDataUtil"},
//	 }
//   NewRequest(args)
func NewRequest(args ...interface{}) (reader io.Reader, header http.Header, err error) {
	var msg AmfMessage
	var msgs []AmfMessage

	switch len(args) {
	case 0:
		err = errors.New("no argument given")
		return
	case 1: // array, multiple requests
		msgs, err = getAmfMessages(args[0])
		if err != nil {
			return
		}
	default: // single request
		msg, err = getAmfMessage(1, args...)
		if err != nil {
			return
		}
		msgs = append(msgs, msg)
	}

	bundle := MessageBundle{
		AmfVersion: 3,
		Messages:   msgs,
	}

	var buf []byte

	buffer := bytes.NewBuffer(buf)
	encoder := NewEncoder(buffer)

	err = EncodeMessageBundle(encoder, &bundle)
	if err != nil {
		return
	}

	header = http.Header{}
	header.Add("Content-Type", "application/x-amf")

	reader = buffer
	return
}

func getAmfMessages(mData interface{}) (msgs []AmfMessage, err error) {
	var msg AmfMessage

	if data, ok := mData.([]interface{}); ok {
		for k, v := range data {
			if single, ok := v.([]interface{}); ok {
				msg, err = getAmfMessage(k+1, single...)
				if err != nil {
					return
				}
				msgs = append(msgs, msg)
			} else {
				err = errors.New("single arguments should be []interface{}")
				return
			}
		}
	} else {
		err = errors.New("getAmfMessages, argument should be []interface{}")
	}
	return
}

func getAmfMessage(msgIdx int, optArgs ...interface{}) (msg AmfMessage, err error) {
	var s string

	targetUri := "null"
	responseUri := fmt.Sprintf("/%d", msgIdx)
	dSEndpoint := ""
	body := []interface{}{}

	optArgsLength := len(optArgs)

	if optArgsLength < 2 {
		err = errors.New("need at least 2 arguments: operation, destination")
		return
	}

	operation, err := getOptArgString(optArgs[0])
	if err != nil {
		return
	}
	destination, err := getOptArgString(optArgs[1])
	if err != nil {
		return
	}

	if optArgsLength > 2 {
		body, err = getOptArgBody(optArgs[2])
		if err != nil {
			return
		}
	}

	if optArgsLength > 3 {
		s, err = getOptArgString(optArgs[3])
		if err != nil {
			return
		}
		targetUri = s
	}

	if optArgsLength > 4 {
		s, err = getOptArgString(optArgs[4])
		if err != nil {
			return
		}
		responseUri = s
	}

	if optArgsLength > 5 {
		s, err = getOptArgString(optArgs[5])
		if err != nil {
			return
		}
		dSEndpoint = s
	}

	msg = AmfMessage{
		TargetUri:   targetUri,
		ResponseUri: responseUri,
		Body: FlexRemotingMessage{
			MessageId:   strings.ToUpper(uuid.New().String()),
			ClientId:    strings.ToUpper(uuid.New().String()),
			Body:        body,
			Operation:   operation,
			Destination: destination,
			Headers: map[string]interface{}{
				"DSId":       strings.ToUpper(uuid.New().String()),
				"DSEndpoint": dSEndpoint,
			},
		},
	}

	return
}

func getOptArgBody(v interface{}) (arr []interface{}, err error) {
	if val, ok := v.([]interface{}); ok {
		arr = val
		return
	}
	err = errors.New("opt arg invalid, should be []interface{}")
	return
}

func getOptArgString(v interface{}) (str string, err error) {
	if s, ok := v.(string); ok {
		str = s
		return
	}
	err = errors.New("opt arg invalid, should be string")
	return
}

// Parse response to [][]map[string]string for body which is an 1-level objects array
//
// 1-level objects array means, each object in the array has no *object* type element[s],
// all elements of this object should be plain type like Integer, String, Number, Null, Date ...
//
// If response body is not array of 1-level objects, you may need to parse it manually
func ParseRespBody(b []byte) (body [][]map[string]string, err error) {
	var bodyElement []map[string]string

	data := bytes.NewBuffer(b)
	bundle, _ := DecodeMessageBundle(data)

	for i := 0; i < len(bundle.Messages); i++ {
		if obj, ok := bundle.Messages[i].Body.(AvmObject); !ok {
			err = errors.New(fmt.Sprintf("convert body to AvmObject: %d", i))
		} else {
			if elements, ok := obj.StaticFields["body"].([]interface{}); !ok {
				err = errors.New(fmt.Sprintf("convert body field to array: %d", i))
			} else {
				for _, e := range elements {
					m, err := parseElement(e)
					if err != nil {
						fmt.Printf("parseElement:%s\n", err.Error())
					} else {
						bodyElement = append(bodyElement, m)
					}
				}
			}

			if err != nil {
				return
			}
			body = append(body, bodyElement)
		}
	}

	return
}

func parseElement(obj interface{}) (m map[string]string, err error) {
	if obj, ok := obj.(AvmObject); !ok {
		err = errors.New("element to map")
	} else {
		m = make(map[string]string)
		for k, v := range obj.StaticFields {
			m[k] = toString(v)
		}
	}
	return
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}

	return fmt.Sprintf("%v", v)
}
