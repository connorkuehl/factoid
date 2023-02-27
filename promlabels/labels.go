package promlabels

type RequestStatus string

const (
	RequestSuccess RequestStatus = "success"
	RequestFail                  = "fail"
	RequestReject                = "reject"
)

type Upstream string

const (
	UpstreamRepo Upstream = "repo"
)
