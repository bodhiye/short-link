package main

type Error interface {
	error
	Status() int
}

type StatusError struct {
	Code int
	Err  error
}


