package dto

type ViewRequestBody struct {
	Namespace string `json:"namespace"`
	Resource  string `json:"resource"`
}

type ViewResponseBody struct {
	Allocated int64 `json:"allocated"`
	Capacity  int64 `json:"capacity"`
	Version   int64 `json:"version"`
}
