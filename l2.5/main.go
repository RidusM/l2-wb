package main

type customError struct {
	msg string
}

func (e *customError) Error() string { // method satisfies the builtin error interface
	return e.msg
}

func test() *customError { // returns typed nil pointer (*customError, nil)
	// ... do something
	return nil
}

func main() {
	var err error   // interface variable (type=nil, data=nil)
	err = test()    // err inside => type = *customError, data = nil
	if err != nil { // true, because interface type field is not nil
		println("error")
		return
	}
	println("ok")
} // prints "error" due to typed nil trap