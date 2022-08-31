package sdp

import (
	"bytes"
	"strings"
	"testing"
)

func TestSDPsParse(t *testing.T) {
	sdpStr := "v=0\r\n" +
		"o=AirTunes 1547303657935225515 0 IN IP4 192.168.0.13\r\n" +
		"s=AirTunes\r\n" +
		"i=iPhone\r\n" +
		"c=IN IP4 192.168.0.13\r\n" +
		"t=0 0\r\n" +
		"m=audio 0 RTP/AVP 96\r\n" +
		"a=rtpmap:96 AppleLossless\r\n" +
		"a=fmtp:96 352 0 16 40 10 14 2 255 0 0 44100\r\n" +
		"a=rsaaeskey:i02+A4Xxkd5GtTdOQqq3xioo3IBHMNRCSAN1B4g5ARI5rYKcO/yFdHNoqXY4P94gvwa9Ofz/cIxaFFsM4PLjXbyrG2/XYdRD6RRi9jyDYHEEE6VAsfsUPptASzKUmlZzwiKnIcR/XE7H6RsafHssJaiBXy+eX7OriRdCivnA8gBl0CgyNvEg5+uOToNJX7E0HOA3ab3cKlxcQeAAPO938mJBbr7tsazvpBYe/70ekjJFaj+HuzFvfbHNbXZMJQfX9tPlHLFgV6xfhkG+ZNX3+av0FNTW3oO1hWzVWElNGsfiV901autNAzQc2v9NvmCXOHB1f/6+pjlapxOU/Br7TQ==\r\n" +
		"a=aesiv:PP+Vz2mBa/rqDGbxvXiXqw==\r\n" +
		"a=min-latency:11025\r\n" +
		"a=max-latency:88200\r\n"

	r := strings.NewReader(sdpStr)
	sdp, err := Parse(r)

	if err != nil {
		t.Error("Expected non nil err value", err)
	}
	if sdp.Version != 0 {
		t.Error("Unexpected version", sdp.Version)
	}
	if sdp.SessionName != "AirTunes" {
		t.Error("Unexpected sessionName", sdp.SessionName)
	}
	if sdp.Information != "iPhone" {
		t.Error("Unexpected information value", sdp.Information)
	}

	o := sdp.Origin
	if o.NetType != "IN" {
		t.Error("Unexpected origin net type", o.NetType)
	}
	if o.AddrType != "IP4" {
		t.Error("Unexpected origin addr type", o.AddrType)
	}
	if o.SessionID != "1547303657935225515" {
		t.Error("Unexpected origin asesion id", o.SessionID)
	}
	if o.SessionVersion != "0" {
		t.Error("Unexpected origin sesion version", o.SessionVersion)
	}
	if o.UnicastAddress != "192.168.0.13" {
		t.Error("Unexpected origin unicast address", o.UnicastAddress)
	}
	if o.Username != "AirTunes" {
		t.Error("Unexpected origin username", o.Username)
	}

	c := sdp.ConnectData
	if c.AddrType != "IP4" {
		t.Error("Unexpected connection addr type", c.AddrType)
	}
	if c.NetType != "IN" {
		t.Error("Unexpected connection net type", c.NetType)
	}
	if c.ConnectionAddress != "192.168.0.13" {
		t.Error("Unexpected connection address", c.ConnectionAddress)
	}

	m := sdp.MediaDescription
	if len(m) != 1 {
		t.Error("Unexpected number of Media Descriptions", len(m))
	}
	if m[0].Media != "audio" {
		t.Error("Unexpected media description media", m[0].Media)
	}
	if m[0].Fmt != "96" {
		t.Error("Unexpected media description format", m[0].Fmt)
	}
	if m[0].Port != "0" {
		t.Error("Unexpected media description port", m[0].Port)
	}
	if m[0].Proto != "RTP/AVP" {
		t.Error("Unexpected media description protocol", m[0].Proto)
	}
	a := sdp.Attributes
	if len(a) != 6 {
		t.Error("Unexpected number of attributes", len(a))
	}
	if a["rtpmap"] != "96 AppleLossless" {
		t.Error("Unexpected rtpmap", a["rtpmap"])
	}
	if a["fmtp"] != "96 352 0 16 40 10 14 2 255 0 0 44100" {
		t.Error("Unexpected fmtp", a["fmtp"])
	}
	if a["rsaaeskey"] != "i02+A4Xxkd5GtTdOQqq3xioo3IBHMNRCSAN1B4g5ARI5rYKcO/yFdHNoqXY4P94gvwa9Ofz/cIxaFFsM4PLjXbyrG2/XYdRD6RRi9jyDYHEEE6VAsfsUPptASzKUmlZzwiKnIcR/XE7H6RsafHssJaiBXy+eX7OriRdCivnA8gBl0CgyNvEg5+uOToNJX7E0HOA3ab3cKlxcQeAAPO938mJBbr7tsazvpBYe/70ekjJFaj+HuzFvfbHNbXZMJQfX9tPlHLFgV6xfhkG+ZNX3+av0FNTW3oO1hWzVWElNGsfiV901autNAzQc2v9NvmCXOHB1f/6+pjlapxOU/Br7TQ==" {
		t.Error("Unexpected rsaaeskey", a["rsaaeskey"])
	}
	if a["aesiv"] != "PP+Vz2mBa/rqDGbxvXiXqw==" {
		t.Error("Unexpected aesiv", a["aesiv"])
	}
	if a["min-latency"] != "11025" {
		t.Error("Unexpected min-latency", a["min-latency"])
	}
	if a["max-latency"] != "88200" {
		t.Error("Unexpected max-latency", a["max-latency"])
	}

}

func TestSDPWrite(t *testing.T) {
	sdpStr := "v=0\r\n" +
		"o=AirTunes 1547303657935225515 0 IN IP4 192.168.0.13\r\n" +
		"s=AirTunes\r\n" +
		"c=IN IP4 192.168.0.13\r\n" +
		"t=0 0\r\n" +
		"m=audio 0 RTP/AVP 96\r\n" +
		"a=rtpmap:96 AppleLossless\r\n" +
		"i=iPhone\r\n"

	session := SessionDescription{}
	session.Version = 0
	session.SessionName = "AirTunes"
	session.Information = "iPhone"
	origin := Origin{}
	origin.Username = "AirTunes"
	origin.SessionID = "1547303657935225515"
	origin.SessionVersion = "0"
	origin.NetType = "IN"
	origin.AddrType = "IP4"
	origin.UnicastAddress = "192.168.0.13"
	session.Origin = origin
	c := ConnectData{}
	c.AddrType = "IP4"
	c.NetType = "IN"
	c.ConnectionAddress = "192.168.0.13"
	session.ConnectData = c
	timing := Timing{StartTime: 0, StopTime: 0}
	session.Timing = timing
	m := make([]MediaDescription, 1)
	md := MediaDescription{}
	md.Media = "audio"
	md.Fmt = "96"
	md.Port = "0"
	md.Proto = "RTP/AVP"
	m[0] = md
	session.MediaDescription = m
	a := make(map[string]string)
	a["rtpmap"] = "96 AppleLossless"
	session.Attributes = a
	var b bytes.Buffer
	n, err := Write(&b, &session)
	if err != nil {
		t.Error("Expected nil err value", err)
	}
	if n <= 0 {
		t.Error("No bytes written")
	}
	if sdpStr != b.String() {
		t.Error("Non matching response generated. Expected:"+sdpStr+"got:", b.String())
	}
}
