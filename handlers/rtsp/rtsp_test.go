package rtsp

import "testing"

func TestMethodExists(t *testing.T) {
	method, err := getMethod("options")
	if err != nil {
		t.Error("Expected nil err value", err)
	}
	if method != Options {
		t.Error("Expected Options, got: ", method)
	}
}

func TestMethodExistsCaseSensitive(t *testing.T) {
	method, err := getMethod("OPTIONS")
	if err != nil {
		t.Error("Expected nil err value", err)
	}
	if method != Options {
		t.Error("Expected Options, got: ", method)
	}
}

func TestMethodNotExists(t *testing.T) {
	_, err := getMethod("foo")
	if err == nil {
		t.Error("Expected non nil err value")
	}

}

func TestStatusExists(t *testing.T) {
	status, err := getStatus(200)
	if err != nil {
		t.Error("Unexpected err value", err)
	}
	if status != Ok {
		t.Error("Expected OK, got: ", status.String())
	}
}

func TestStatusDoesNotExist(t *testing.T) {
	_, err := getStatus(900)
	if err == nil {
		t.Error("Expected error value")
	}
}
