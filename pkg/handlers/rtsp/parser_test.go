package rtsp

import (
	"bytes"
	"strings"
	"testing"
)

func TestOptionsParse(t *testing.T) {
	options :=
		"OPTIONS * RTSP/1.0\r\n" +
			"CSeq: 1\r\n" +
			"User-Agent: iTunes/12.5.1 (Macintosh; OS X 10.11.6)\r\n" +
			"Client-Instance: 67F67C1CAA66A2F4\r\n" +
			"DACP-ID: 67F67C1CAA66A2F4\r\n" +
			"Active-Remote: 1721127963\r\n" +
			"\r\n"

	r := strings.NewReader(options)
	msg, err := readRequest(r)

	if err != nil {
		t.Error("Expected non nil err value", err)
	}
	if msg.Method != Options {
		t.Error("Expected OPTIONS got: ", msg.Method)
	}
	if msg.protocol != "RTSP/1.0" {
		t.Error("Expected RTSP/1.0 got: ", msg.protocol)
	}
	if len(msg.Headers) != 5 {
		t.Error("Unexpected amount of headers: ", len(msg.Headers))
	}
	// test a couple of the headers
	if msg.Headers["CSeq"] != "1" {
		t.Error("Unexpected CSeq", msg.Headers["CSeq"])
	}
	if msg.Headers["Client-Instance"] != "67F67C1CAA66A2F4" {
		t.Error("Unexpected Client-Instance", msg.Headers["Client-Instance"])
	}

}

func TestAnnounceParse(t *testing.T) {

	body := "v=0\r\n" +
		"o=AirTunes 2699324803567405959 0 IN IP4 192.168.1.5\r\n" +
		"s=AirTunes\r\n" +
		"c=IN IP4 192.168.1.5\r\n" +
		"t=0 0\r\n" +
		"m=audio 0 RTP/AVP 96\r\n" +
		"a=rtpmap:96 mpeg4-generic/44100/2\r\n" +
		"a=fmtp:96\r\n" +
		"a=fpaeskey:RlBMWQECAQAAAAA8AAAAAOG6c4aMdLkXAX+lbjp7EhgAAAAQeX5uqGyYkBmJX+gd5ANEr+amI8urqFmvcNo87pR0BXGJ4eLf\r\n" +
		"a=aesiv:VZTaHn4wSJ84Jjzlb94m0Q==\r\n" +
		"a=min-latency:11025\r\n"

	announce := "ANNOUNCE rtsp://192.168.1.45/2699324803567405959 RTSP/1.0\r\n" +
		"X-Apple-Device-ID: 0xa4d1d2800b68\r\n" +
		"CSeq: 16\r\n" +
		"DACP-ID: 14413BE4996FEA4D\r\n" +
		"Active-Remote: 2543110914\r\n" +
		"Content-Type: application/sdp\r\n" +
		"Content-Length: 331\r\n" +
		"\r\n" + body

	r := strings.NewReader(announce)
	msg, err := readRequest(r)

	if err != nil {
		t.Error("Expected non nil err value", err)
	}
	if msg.Method != Announce {
		t.Error("Expected Announce got: ", msg.Method)
	}

	if string(msg.Body) != body {
		t.Error("Expected " + body + " got: " + string(msg.Body))
	}

}

func TestParseImproperRequestLine(t *testing.T) {
	options :=
		"OPTIONS *\r\n" +
			"CSeq: 1\r\n" +
			"User-Agent: iTunes/12.5.1 (Macintosh; OS X 10.11.6)\r\n" +
			"Client-Instance: 67F67C1CAA66A2F4\r\n" +
			"DACP-ID: 67F67C1CAA66A2F4\r\n" +
			"Active-Remote: 1721127963\r\n" +
			"\r\n"

	r := strings.NewReader(options)
	_, err := readRequest(r)
	if err == nil {
		t.Error("Expected error ")
	}

}

func TestParseImproperHeader(t *testing.T) {
	options :=
		"OPTIONS * RTSP/1.0\r\n" +
			"CSeq: 1\r\n" +
			"User-Agent\r\n" +
			"Client-Instance: 67F67C1CAA66A2F4\r\n" +
			"DACP-ID: 67F67C1CAA66A2F4\r\n" +
			"Active-Remote: 1721127963\r\n" +
			"\r\n"

	r := strings.NewReader(options)
	_, err := readRequest(r)
	if err == nil {
		t.Error("Expected non nil err value", err)
	}
}

func TestBuildResponse(t *testing.T) {
	respString :=
		"RTSP/1.0 200 Ok\r\n" +
			"Client-Instance: 67F67C1CAA66A2F4\r\n" +
			"\r\n"
	resp := Response{}
	headers := make(map[string]string)
	headers["Client-Instance"] = "67F67C1CAA66A2F4"
	resp.protocol = "RTSP/1.0"
	resp.Headers = headers
	resp.Status = Ok
	var b bytes.Buffer
	n, err := writeResponse(&b, &resp)
	if err != nil {
		t.Error("Unexpected err value", err)
	}
	if n <= 0 {
		t.Error("No bytes written")
	}
	if respString != b.String() {
		t.Error("Non matching response generated. Expected:"+respString+"got:", b.String())
	}

}

func TestWriteRequest(t *testing.T) {
	requestStr :=
		"OPTIONS * RTSP/1.0\r\n" +
			"Client-Instance: 67F67C1CAA66A2F4\r\n" +
			"\r\n"
	request := Request{}
	request.Method = Options
	request.protocol = "RTSP/1.0"
	request.RequestURI = "*"
	headers := make(map[string]string)
	headers["Client-Instance"] = "67F67C1CAA66A2F4"
	request.Headers = headers
	var b bytes.Buffer
	n, err := writeRequest(&b, &request)
	if err != nil {
		t.Error("Expected nil err value", err)
	}
	if n <= 0 {
		t.Error("No bytes written")
	}
	if requestStr != b.String() {
		t.Error("Non matching response generated. Expected:"+requestStr+"got:", b.String())
	}
}

func TestWriteRequestBody(t *testing.T) {
	requestStr :=
		"OPTIONS * RTSP/1.0\r\n" +
			"Content-Length: 4\r\n" +
			"\r\n" +
			"abcd"
	request := Request{}
	request.Method = Options
	request.protocol = "RTSP/1.0"
	request.RequestURI = "*"
	var bodyBuffer bytes.Buffer
	bodyBuffer.WriteString("a")
	bodyBuffer.WriteString("b")
	bodyBuffer.WriteString("c")
	bodyBuffer.WriteString("d")
	request.Body = bodyBuffer.Bytes()
	var b bytes.Buffer
	n, err := writeRequest(&b, &request)
	if err != nil {
		t.Error("Expected nil err value", err)
	}
	if n <= 0 {
		t.Error("No bytes written")
	}
	if requestStr != b.String() {
		t.Error("Non matching response generated. Expected:"+requestStr+"got:", b.String())
	}
}

func TestParseResponse(t *testing.T) {
	responseString := "RTSP/1.0 200 OK\r\n" +
		"Public: ANNOUNCE, SETUP, RECORD, PAUSE, FLUSH, TEARDOWN, OPTIONS,GET_PARAMETER, SET_PARAMETER, POST, GET\r\n" +
		"Server: AirTunes/130.14\r\n" +
		"CSeq: 3\r\n" +
		"\r\n"
	r := strings.NewReader(responseString)
	resp, err := readResponse(r)
	if err != nil {
		t.Error("Unexpected err value", err)
	}
	if resp.Status != Ok {
		t.Error("Non matching status generated. Expected:"+Ok.String()+"got:", resp.Status.String())
	}
	if resp.protocol != "RTSP/1.0" {
		t.Error("Non matching protocol generated. Expected:"+"RTSP/1.0"+"got:", resp.protocol)
	}
}
