package main

type requestStatus string

const (
	requestSuccess requestStatus = "success"
	requestFail                  = "fail"
	requestReject                = "reject"
)
