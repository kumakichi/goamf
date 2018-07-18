package amf

import (
	"errors"
	"io"
	"strings"
)

// function for WIP code:
func unused(a ...interface{}) {}

var STATUS_CODES = map[string]string{
	"STATUS_OK":    "/onResult",
	"STATUS_ERROR": "/onStatus",
	"STATUS_DEBUG": "/onDebugEvents",
}

type FlexRemotingMessage struct {
	// AbstractMessage:
	Body        []interface{}
	ClientId    string
	Destination string
	Headers     map[string]interface{}
	MessageId   string
	Timestamp   uint32
	TimeToLive  uint32

	// RemotingMessage:
	Operation string
	Source    string
}

type FlexErrorMessage struct {
	// AbstractMessage:
	Body        []interface{}
	ClientId    string
	Destination string
	Headers     map[string]interface{}
	MessageId   string
	Timestamp   uint32
	TimeToLive  uint32

	// AcknowledgeMessage:
	Flags byte

	// ErrorMessage:
	ExtendedData string
	FaultCode    int
	FaultDetail  int
	FaultString  string
	RootCause    string
}

type MessageBundle struct {
	AmfVersion uint16
	Headers    []Header
	Messages   []AmfMessage
}

type Header struct {
	Name           string
	MustUnderstand bool
	Value          interface{}
}
type AmfMessage struct {
	TargetUri   string
	ResponseUri string
	Body        interface{}
}

func DecodeMessageBundle(stream io.Reader) (*MessageBundle, error) {

	cxt := NewDecoder(stream, 0)
	cxt.RegisterType("flex.messaging.messages.RemotingMessage", FlexRemotingMessage{})

	amfVersion := cxt.ReadUint16()

	result := MessageBundle{}
	cxt.AmfVersion = amfVersion
	result.AmfVersion = amfVersion

	/*
	   From http://osflash.org/documentation/amf/envelopes/remoting:

	   The first two bytes of an AMF message are an unsigned short int. The result
	   indicates what type of Flash Player connected to the server.

	   0x00 for Flash Player 8 and below
	   0x01 for FlashCom/FMS
	   0x03 for Flash Player 9
	   Note that Flash Player 9 will always set the second byte to 0x03, regardless of
	   whether the message was sent in AMF0 or AMF3.
	*/

	if cxt.AmfVersion > 0x09 {
		return nil, errors.New("Malformed stream (wrong amfVersion)")
	}

	headerCount := cxt.ReadUint16()

	/*
	   From http://osflash.org/documentation/amf/envelopes/remoting:

	   Each header consists of the following:

	   UTF string (including length bytes) - name
	   Boolean - specifies if understanding the header is 'required'
	   Long - Length in bytes of header
	   Variable - Actual data (including a type code)
	*/

	// Read headers
	result.Headers = make([]Header, headerCount)
	for i := 0; i < int(headerCount); i++ {
		_, name := cxt.ReadString()
		mustUnderstand := cxt.ReadUint8() != 0
		data_len := cxt.ReadUint32()
		unused(data_len)

		value := cxt.ReadValue()
		header := Header{name, mustUnderstand, value}
		result.Headers[i] = header

		//fmt.Printf("Read header, name = %s\n", name)
	}

	/*
	   From http://osflash.org/documentation/amf/envelopes/remoting:

	   Between the headers and the start of the bodies is a int specifying the number of
	   bodies. Each body consists of the following:

	   UTF String - Target
	   UTF String - Response
	   Long - Body length in bytes
	   Variable - Actual data (including a type code)
	*/

	// Read message bodies
	messageCount := cxt.ReadUint16()
	result.Messages = make([]AmfMessage, messageCount)

	for i := 0; i < int(messageCount); i++ {
		// TODO: Should reset object tables here
		cxt.Clear()

		message := &result.Messages[i]

		_, message.TargetUri = cxt.ReadString()
		_, message.ResponseUri = cxt.ReadString()

		status := "STATUS_OK"
		is_request := true
		for code, s := range STATUS_CODES {
			if !strings.HasSuffix(message.TargetUri, s) {
				continue
			}
			is_request = false
			status = code
			message.TargetUri = message.TargetUri[len(message.TargetUri)-len(s)+1 : len(message.TargetUri)]
			_ = status
		}

		//fmt.Println(message)
		messageLength := cxt.ReadUint32()
		// TODO: Check targetUri to see if this isn't an array?

		if is_request {
			// Read an array, however this array is strange because it doesn't use
			// the reference bit.
			typeCode := cxt.ReadUint8()
			if typeCode != 9 {
				return nil, errors.New("Expected Array type code in message body")
			}
			ref := cxt.ReadUint32()
			itemCount := int(ref)
			args := make([]interface{}, itemCount)
			for i := 0; i < itemCount; i++ {
				args[i] = cxt.ReadValue()
			}
			message.Body = args
		} else {
			message.Body = cxt.ReadValue()
		}

		unused(messageLength)
	}

	return &result, nil
}

// Encode message for http request
func EncodeMessageBundle(cxt *Encoder, bundle *MessageBundle) error {
	cxt.WriteUint16(bundle.AmfVersion)

	// Write headers
	cxt.WriteUint16(uint16(len(bundle.Headers)))
	for _, header := range bundle.Headers {
		cxt.WriteString(header.Name)
		cxt.WriteBool(header.MustUnderstand)
	}

	// Write messages
	cxt.WriteUint16(uint16(len(bundle.Messages)))
	for _, message := range bundle.Messages {
		cxt.WriteString(message.TargetUri)
		cxt.WriteString(message.ResponseUri)
		cxt.WriteUint32(0)

		cxt.WriteUint8(0x0a)
		cxt.WriteUint32(0x01) // message body length
		cxt.WriteUint8(0x11)  // type amf3
		cxt.WriteValueAmf3(message.Body)
	}

	return nil
}
