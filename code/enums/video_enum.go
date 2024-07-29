package enums

type VideoType string

const (
	Video_FLV  VideoType = ".flv"
	Video_MP4  VideoType = ".mp4"
	Video_MKV  VideoType = ".mkv"
	Video_MOV  VideoType = ".mov"
	Video_WMV  VideoType = ".wmv"
	Video_AVI  VideoType = ".avi"
	Video_RM   VideoType = ".rm"
	Video_RMVB VideoType = ".rmvb"
	Video_ASF  VideoType = ".asf"
	Video_MTS  VideoType = ".mts"
)

// func (v VideoType) String() string {
// 	switch (v){
// 		case Video_FLV: return
// 	}
// }

func (v VideoType) GetAll() []VideoType {
	return []VideoType{
		Video_FLV,
		Video_MP4,
		Video_MKV,
		Video_MOV,
		Video_WMV,
		Video_AVI,
		Video_RM,
		Video_RMVB,
		Video_ASF,
		Video_MTS,
	}
}

func (v VideoType) GetAllString() []string {
	return []string{
		string(Video_FLV),
		string(Video_MP4),
		string(Video_MKV),
		string(Video_MOV),
		string(Video_WMV),
		string(Video_AVI),
		string(Video_RM),
		string(Video_RMVB),
		string(Video_ASF),
		string(Video_MTS),
	}
}
