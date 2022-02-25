package statpoll

type ReqType string

const (
	ReqBridgeInfo     ReqType = "bridge_info"
	ReqStopBridgeInfo ReqType = "stop_bridge_info"

	ReqYoutubeInfo     ReqType = "youtube_info"
	ReqStopYoutubeInfo ReqType = "stop_youtube_info"

	ReqAirPlayInfo     ReqType = "airplay_info"
	ReqStopAirPlayInfo ReqType = "stop_airplay_info"
)

type ReqParam string

const (
	YtParamNowPlaying    ReqParam = "yt_now_playing"
	YtParamTrackDuration ReqParam = "yt_track_duration"
	YtParamElapsedTime   ReqParam = "yt_elapsed_time"
	YtParamPaused        ReqParam = "yt_paused"
	YtParamTrackIndex    ReqParam = "yt_track_index"
	YtParamQueuedTracks  ReqParam = "yt_queued_tracks"
)
