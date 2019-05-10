package main

import (
	"context"
	"os"

	"github.com/m-mizutani/deepalert"
)

func dummySearcher(ctx context.Context, attr deepalert.Attribute) ([]deepalert.Attribute, error) {
	newAttr := deepalert.Attribute{
		Key:   "username",
		Value: "mizutani",
		Type:  deepalert.TypeUserName,
	}
	return []deepalert.Attribute{newAttr}, nil
}

func main() {
	deepalert.StartSearch(dummySearcher, "dummySearcher", os.Getenv("ATTRIBUTE_TOPIC"))
}
