package promlabels

type RequestStatus string

func (r RequestStatus) String() string {
	return string(r)
}

const (
	RequestSuccess RequestStatus = "success"
	RequestFail    RequestStatus = "fail"
	RequestReject  RequestStatus = "reject"
)

type Upstream string

func (u Upstream) String() string {
	return string(u)
}

const (
	UpstreamRepo Upstream = "repo"
)
