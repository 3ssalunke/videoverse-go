package utils

type VideoTrimRequest struct {
	VideoID string  `json:"video_id"`
	StartTS float64 `json:"start_ts"`
	EndTS   float64 `json:"end_ts"`
}
