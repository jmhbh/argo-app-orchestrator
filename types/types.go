package types

type LoggerKey struct{}

type UserMetadata struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// PodMetadata currently unused but could be useful to replace kickChan in the future
type PodMetadata struct {
	Name string `json:"name"`
}

type Params struct {
	UserMetadataChan chan UserMetadata
	KickChan         chan struct{}
}
