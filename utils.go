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

func NewRequest(operation, destination string, optArgs ...string) (reader io.Reader, header http.Header, err error) {
	var buf []byte
	buffer := bytes.NewBuffer(buf)
	encoder := NewEncoder(buffer)

	targetUri := "null"
	responseUri := "/1"
	dSEndpoint := ""

	optsLength := len(optArgs)
	switch optsLength {
	case 1:
		targetUri = optArgs[0]
	case 2:
		targetUri = optArgs[0]
		responseUri = optArgs[1]
	case 3:
		targetUri = optArgs[0]
		responseUri = optArgs[1]
		dSEndpoint = optArgs[2]
	}

	bundle := MessageBundle{
		AmfVersion: 3,
		Messages: []AmfMessage{
			{
				TargetUri:   targetUri,
				ResponseUri: responseUri,
				Body: FlexRemotingMessage{
					MessageId:   strings.ToUpper(uuid.New().String()),
					ClientId:    strings.ToUpper(uuid.New().String()),
					Operation:   operation,
					Destination: destination,
					Headers: map[string]interface{}{
						"DSId":       strings.ToUpper(uuid.New().String()),
						"DSEndpoint": dSEndpoint,
					},
				},
			},
		},
	}

	err = EncodeMessageBundle(encoder, &bundle)
	if err != nil {
		return
	}

	header = http.Header{}
	header.Add("Content-Type", "application/x-amf")
	header.Add("Accept-Encoding", "identity")
	header.Add("Connection", "close")

	reader = buffer
	return
}

func ParseRespBody(b []byte) (body []map[string]string, err error) {
	data := bytes.NewBuffer(b)
	bundle, _ := DecodeMessageBundle(data)

	if obj, ok := bundle.Messages[0].Body.(AvmObject); !ok {
		err = errors.New("convert to AvmObject")
	} else {
		if elements, ok := obj.StaticFields["body"].([]interface{}); !ok {
			err = errors.New("convert body field to array")
		} else {
			for _, e := range elements {
				m, err := parseElement(e)
				if err != nil {
					fmt.Printf("parseElement:%s\n", err.Error())
				} else {
					body = append(body, m)
				}
			}
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
