package b3

// with a go.mod i'm pretty sure we don't want or need GOPATH set. With GOPATH set, go test says
// "$GOPATH/go.mod exists but should not"

func Hello() string {
	return "Hello, world."
}



