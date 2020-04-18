package main

import (
	"bufio"
	"os"

	"github.com/deiu/rdf2go"
	// "github.com/k0kubun/pp"
)

func main() {

	// Set a base URI
	uri := "https://www.cars.com/vehicledetail/detail/804122829/overview/"

	// Check remote server certificate to see if it's valid 
	// (don't skip verification)
	skipVerify := false

	// Create a new graph. You can also omit the skipVerify parameter
	// and accept invalid certificates (e.g. self-signed)
	g := rdf2go.NewGraph(uri, skipVerify)

	err := g.LoadURI(uri)
	if err != nil {
		// deal with the error
	}

   	f, err := os.Create("turtle.txt")
	if err != nil {
		// deal with the error
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	// w is of type io.Writer
	g.Serialize(w, "text/turtle")
	w.Flush()

}
