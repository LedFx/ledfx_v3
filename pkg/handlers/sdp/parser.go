package sdp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parse parses out an SDP packet into a SDP struct
func Parse(r io.Reader) (*SessionDescription, error) {
	sdp := NewSessionDescription()
	s := bufio.NewScanner(r)
	for s.Scan() {
		parts := strings.SplitN(s.Text(), "=", 2)
		typePart := parts[0]
		valuePart := parts[1]
		switch typePart {
		case "v":
			version, err := strconv.Atoi(valuePart)
			if err != nil {
				return nil, err
			}
			sdp.Version = version
		case "o":
			// <username> <sess-id> <sess-version> <nettype> <addrtype> <unicast-address>
			originParts := strings.Fields(valuePart)
			origin := Origin{}
			origin.Username = originParts[0]
			origin.SessionID = originParts[1]
			origin.SessionVersion = originParts[2]
			origin.NetType = originParts[3]
			origin.AddrType = originParts[4]
			origin.UnicastAddress = originParts[5]
			sdp.Origin = origin
		case "s":
			sdp.SessionName = valuePart
		case "c":
			// <nettype> <addrtype> <connection-address>
			connectionParts := strings.Fields(valuePart)
			connect := ConnectData{}
			connect.NetType = connectionParts[0]
			connect.AddrType = connectionParts[1]
			connect.ConnectionAddress = connectionParts[2]
			sdp.ConnectData = connect
		case "t":
			// <start-time> <stop-time>
			timingParts := strings.Fields(valuePart)
			timing := Timing{}
			start, err := strconv.Atoi(timingParts[0])
			if err != nil {
				return nil, err
			}
			stop, err := strconv.Atoi(timingParts[1])
			if err != nil {
				return nil, err
			}
			timing.StartTime = start
			timing.StopTime = stop
			sdp.Timing = timing
		case "m":
			// <media> <port>/<number of ports> <proto> <fmt>
			mediaParts := strings.Fields(valuePart)
			media := MediaDescription{}
			media.Media = mediaParts[0]
			media.Port = mediaParts[1]
			media.Proto = mediaParts[2]
			media.Fmt = mediaParts[3]
			sdp.MediaDescription = append(sdp.MediaDescription, media)
		case "a":
			attributeParts := strings.Split(valuePart, ":")
			sdp.Attributes[attributeParts[0]] = attributeParts[1]
		case "i":
			sdp.Information = valuePart
		}
		//TODO: handle all parameters

	}
	return sdp, nil
}

//Write writes a SessionDescription struct to the given writer
func Write(w io.Writer, session *SessionDescription) (n int, err error) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("v=%d\r\n", session.Version))
	// <username> <sess-id> <sess-version> <nettype> <addrtype> <unicast-address>
	o := session.Origin
	buf.WriteString(fmt.Sprintf("o=%s %s %s %s %s %s\r\n", o.Username, o.SessionID, o.SessionVersion, o.NetType, o.AddrType, o.UnicastAddress))
	buf.WriteString(fmt.Sprintf("s=%s\r\n", session.SessionName))
	// <nettype> <addrtype> <connection-address>
	c := session.ConnectData
	buf.WriteString(fmt.Sprintf("c=%s %s %s\r\n", c.NetType, c.AddrType, c.ConnectionAddress))
	// <start-time> <stop-time>
	t := session.Timing
	buf.WriteString(fmt.Sprintf("t=%d %d\r\n", t.StartTime, t.StopTime))
	for _, m := range session.MediaDescription {
		// <media> <port>/<number of ports> <proto> <fmt>
		buf.WriteString(fmt.Sprintf("m=%s %s %s %s\r\n", m.Media, m.Port, m.Proto, m.Fmt))
	}
	for k, v := range session.Attributes {
		buf.WriteString(fmt.Sprintf("a=%s:%s\r\n", k, v))
	}
	buf.WriteString(fmt.Sprintf("i=%s\r\n", session.Information))
	return w.Write(buf.Bytes())
}
