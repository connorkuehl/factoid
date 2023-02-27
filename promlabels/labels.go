package promlabels

type RequestStatus string

const (
	RequestSuccess RequestStatus = "success"
	RequestFail                  = "fail"
	RequestReject                = "reject"
)
