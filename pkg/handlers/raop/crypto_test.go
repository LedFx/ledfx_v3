package raop

import (
	"net"
	"testing"
)

func TestResponseGenerate(t *testing.T) {
	expectedResp := "r89JJyLNRJ0RT/pI7OqyDzyF0ggoUY0BmpFB9hsIDkziT+TYZ6coZwdBX8AQWQiNGYQBSNzcFWQj41kGcUGOhE2OxnphwHjraZRvF5bwvcvjKEFmkJTtEDnfLvYB41MfzTbWDWA3PSXxVkOrfnMb0hRnS6Es4WWfuSzDDRKQBQUUvob4mrHh9QuMYU+uTbOEE8zXY4QWAjQuOJH8vPSyUmonJLRRdtftgMqxfRjPEJV+4XuZ5vv347ahg3Yr8K12kKJ7axyrJVbF6ghkkCM64Xn6iD6x7p453VjS5gtuz8pLECidA8yudBdJPIASAIRNownnuL/7GQy1bmRIFDvhsw"
	mac, _ := net.ParseMAC("54:52:00:b8:58:77")
	resp, _ := generateChallengeResponse("gY3cmhtK9LnECNUlXFb0qg==", mac, "192.168.0.15")
	if resp != expectedResp {
		t.Errorf("Expected: %s\r\n Got: %s", expectedResp, resp)
	}
}

func TestResponseGenerateIPv6(t *testing.T) {
	expectedResp := "OVq+aJeTOvhFEItbsHrEp82mCvbbC8Nlw6CmSGfEW1LfPWJ0C4asxzl3kSJvy1SzvWZII0oHq18mAsv0ycF3B+tWKrc9TOzng9kyQvzKTwqjscUjjqh0x/m6kedetJ7vIGxD8JbdaG5W7oN8f0IIgHRcXcNfw1wZ5EctlTjkBypXFJN+bgQgie+f8N+ui3WaSp6/sFSdZV820kNW8OqQItqEVZPz199TFwxMYGqJBBC62pbZlV1qoFTiPhDIcIqLiDHHvSIj3b9uFaYA2juVx1YCcbsJ9EsKTItIP3ONgoLDFf+VC0BBSIylQ2fJ/4L0CxMdiTUW3YeMw3WYmHtIMQ"
	mac, _ := net.ParseMAC("04:0c:ce:df:c6:d8")
	resp, _ := generateChallengeResponse("4nQ5iywx/G99yNw9f6oPPg==", mac, "fe80::60c:ceff:fedf:c6d8")
	if resp != expectedResp {
		t.Errorf("Expected: %s\r\n Got: %s", expectedResp, resp)
	}
}
