package statpoll

type ReqType string

const (
	RqtBridgeInfo     ReqType = "bridge_info"
	RqtStopBridgeInfo ReqType = "stop_bridge_info"

	RqtYoutubeInfo     ReqType = "youtube_info"
	RqtStopYoutubeInfo ReqType = "stop_youtube_info"

	RqtAirPlayInfo     ReqType = "airplay_info"
	RqtStopAirPlayInfo ReqType = "stop_airplay_info"
)
