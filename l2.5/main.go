package main

type customError struct {
	msg string
}

func (e *customError) Error() string { // this method need to correct work with customError
	return e.msg
}

func test() *customError { // return value is customError
	// ... do something
	return nil
}

func main() {
	var err error
	err = test()    // err => data = nil, type = *customError
	if err != nil { // err != nil, cuz err type is not nil | data and type should be nil
		println("error")
		return
	}
	println("ok")
} // return error