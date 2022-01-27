package rtsp

import (
	"bytes"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
	"strings"
)

// Request RTSP request
type Request struct {
	Method     Method
	RequestURI string
	protocol   string
	Headers    map[string]string
	Body       []byte
}

// Response RTSP response
type Response struct {
	Headers  map[string]string
	Body     []byte
	Status   Status
	protocol string
}

func NewResponse() *Response {
	return &Response{Headers: make(map[string]string)}
}

func NewRequest() *Request {
	return &Request{Headers: make(map[string]string), protocol: "RTSP/1.0"}
}

func (r *Request) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Protocol: %s\nMethod: %s\nRequest URI: %s\n", r.protocol, r.Method.String(), r.RequestURI))
	buffer.WriteString("Headers:\n")
	for k, v := range r.Headers {
		buffer.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	if contentType, ok := r.Headers["Content-Type"]; ok {
		switch {
		case fastContains(contentType, "image"):
			fallthrough
		case fastContains(contentType, "x-dmap-tagged"):
			buffer.WriteString("Body: <omitted due to length>")
			return buffer.String()
		default:
			buffer.WriteString(fmt.Sprintf("Body: \n %s", r.Body))
		}
	}
	return buffer.String()
}

func (r *Response) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Protocol: %s\nStatus: %s\n", r.protocol, r.Status.String()))
	buffer.WriteString("Headers:\n")
	for k, v := range r.Headers {
		buffer.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	if contentType, ok := r.Headers["Content-Type"]; ok {
		switch {
		case fastContains(contentType, "image"):
			fallthrough
		case fastContains(contentType, "x-dmap-tagged"):
			buffer.WriteString("Body: <omitted due to length>")
			return buffer.String()
		default:
			buffer.WriteString(fmt.Sprintf("Body: \n %s", r.Body))
		}
	}
	return buffer.String()
}

var methods = map[string]Method{
	strings.ToLower(Describe.String()):      Describe,
	strings.ToLower(Announce.String()):      Announce,
	strings.ToLower(Get_Parameter.String()): Get_Parameter,
	strings.ToLower(Options.String()):       Options,
	strings.ToLower(Play.String()):          Play,
	strings.ToLower(Pause.String()):         Pause,
	strings.ToLower(Record.String()):        Record,
	strings.ToLower(Setup.String()):         Setup,
	strings.ToLower(Set_Parameter.String()): Set_Parameter,
	strings.ToLower(Teardown.String()):      Teardown,
	strings.ToLower(Flush.String()):         Flush,
}

// getMethod converts string to Method enum value, returning error if it can't map
func getMethod(method string) (Method, error) {
	m, exists := methods[strings.ToLower(method)]
	if !exists {
		return -1, fmt.Errorf("Not valid method: %s", method)
	}
	return m, nil
}

// GetMethods all RTSP methods as a slice of strings
func GetMethods() []string {
	keys := make([]string, 0, len(methods))
	for k := range methods {
		keys = append(keys, strings.ToUpper(k))
	}
	return keys
}

var statuses = map[int]Status{
	100: Continue,
	200: Ok,
	201: Created,
	250: LowOnStorage,
	300: MultipleChoices,
	301: MovedPermanently,
	303: SeeOther,
	305: UseProxy,
	400: BadRequest,
	401: Unauthorized,
	402: PaymentRequired,
	403: Forbidden,
	404: NotFound,
	405: MethodNotAllowed,
	406: NotAcceptable,
	407: ProxyAuthenticationRequired,
	408: RequestTimeout,
	410: Gone,
	411: LengthRequired,
	412: PreconditionFailed,
	413: RequestEntityTooLarge,
	414: RequestURITooLong,
	415: UnsupportedMediaType,
	451: Invalidparameter,
	452: IllegalConferenceIdentifier,
	453: NotEnoughBandwidth,
	454: SessionNotFound,
	455: MethodNotValidInThisState,
	456: HeaderFieldNotValid,
	457: InvalidRange,
	458: ParameterIsReadOnly,
	459: AggregateOperationNotAllowed,
	460: OnlyAggregateOperationAllowed,
	461: UnsupportedTransport,
	462: DestinationUnreachable,
	500: InternalServerError,
	501: NotImplemented,
	502: BadGateway,
	503: ServiceUnavailable,
	504: GatewayTimeout,
	505: RTSPVersionNotSupported,
	551: Optionnotsupport,
}

func getStatus(status int) (Status, error) {
	s, exists := statuses[status]
	if !exists {
		return -1, fmt.Errorf("Not valid status: %d", status)
	}
	return s, nil
}

func fastContains(str string, substr string) bool {

	start, _ := search.New(language.English, search.IgnoreCase).IndexString(str, substr)
	return start != -1
}
