package util

import (
	"net/http"

	"github.com/LedFx/ledfx/pkg/logger"
)

func InternalError(context string, err error, writer http.ResponseWriter) bool {
	if err == nil {
		return false
	}
	writer.WriteHeader(http.StatusInternalServerError)
	writer.Write([]byte(err.Error()))
	logger.Logger.WithField("context", context).Error(err)
	return true
}

func BadRequest(context string, err error, writer http.ResponseWriter) bool {
	if err == nil {
		return false
	}
	writer.WriteHeader(http.StatusBadRequest)
	writer.Write([]byte(err.Error()))
	logger.Logger.WithField("context", context).Warn(err)
	return true
}
