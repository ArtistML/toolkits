package grpc

func PanicWhenError(err error) {
	if err != nil {
		panic(err)
	}
}
