package enums

type DefaultVideo string

const (
	DefaultVideo_1920 DefaultVideo = "1920"
	DefaultVideo_1280 DefaultVideo = "1280"
)

func (v DefaultVideo) GetAll() []DefaultVideo {
	return []DefaultVideo{
		DefaultVideo_1920,
		DefaultVideo_1280,
	}
}

func (v DefaultVideo) GetAllString() []string {
	return []string{
		string(DefaultVideo_1920),
		string(DefaultVideo_1280),
	}
}
