//go:generate stringer -type=Status status.go

package rtsp

type Status int

const (
	Continue                      Status = 100
	Ok                            Status = 200
	Created                       Status = 201
	LowOnStorage                  Status = 250
	MultipleChoices               Status = 300
	MovedPermanently              Status = 301
	MovedTemp                     Status = 301
	SeeOther                      Status = 303
	UseProxy                      Status = 305
	BadRequest                    Status = 400
	Unauthorized                  Status = 401
	PaymentRequired               Status = 402
	Forbidden                     Status = 403
	NotFound                      Status = 404
	MethodNotAllowed              Status = 405
	NotAcceptable                 Status = 406
	ProxyAuthenticationRequired   Status = 407
	RequestTimeout                Status = 408
	Gone                          Status = 410
	LengthRequired                Status = 411
	PreconditionFailed            Status = 412
	RequestEntityTooLarge         Status = 413
	RequestURITooLong             Status = 414
	UnsupportedMediaType          Status = 415
	Invalidparameter              Status = 451
	IllegalConferenceIdentifier   Status = 452
	NotEnoughBandwidth            Status = 453
	SessionNotFound               Status = 454
	MethodNotValidInThisState     Status = 455
	HeaderFieldNotValid           Status = 456
	InvalidRange                  Status = 457
	ParameterIsReadOnly           Status = 458
	AggregateOperationNotAllowed  Status = 459
	OnlyAggregateOperationAllowed Status = 460
	UnsupportedTransport          Status = 461
	DestinationUnreachable        Status = 462
	InternalServerError           Status = 500
	NotImplemented                Status = 501
	BadGateway                    Status = 502
	ServiceUnavailable            Status = 503
	GatewayTimeout                Status = 504
	RTSPVersionNotSupported       Status = 505
	Optionnotsupport              Status = 551
)
