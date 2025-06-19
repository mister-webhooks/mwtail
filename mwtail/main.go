// Package main contains an example client for Mister Webhooks.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/davecgh/go-spew/spew"
	"github.com/mister-webhooks/mister-webhooks-client/golang/client"
)

func main() {
	helpFlag := flag.Bool("help", false, "show this usage message")

	noColorFlag := flag.Bool("no-color", false, "whether to disable color output")
	_, noColorEnv := os.LookupEnv("NO_COLOR")

	color := !(noColorEnv || *noColorFlag)

	testModeFlag := flag.Bool("test", false, "run in output test mode")
	listStylesFlag := flag.Bool("list-styles", false, "show all supported colorization styles")
	styleFlag := flag.String("style", "friendly", "the colorization style to use")

	//
	// Parse and validate arguments from the command line
	//
	flag.Parse()

	if *listStylesFlag {
		fmt.Println("supported styles:")
		for _, name := range styles.Names() {
			fmt.Printf("  %s\n", name)
		}
		os.Exit(0)
	}

	if _, supported := styles.Registry[*styleFlag]; !supported {
		fmt.Printf("error: %s is not a supported style\n", *styleFlag)
		os.Exit(1)
	}

	//
	// Configure prettyprinter
	//
	spew.Config.Indent = "  "
	spew.Config.SortKeys = true

	var render func(w io.Writer, a ...interface{}) error
	render = func(w io.Writer, a ...any) error {
		_, err := w.Write([]byte(spew.Sdump(a...)))
		return err
	}

	if color {
		render = func(w io.Writer, a ...any) error {
			return quick.Highlight(
				w,
				spew.Sdump(a...),
				"go",
				"terminal16",
				*styleFlag,
			)
		}
	}

	//
	// Run test mode
	//
	if *testModeFlag {
		err := render(log.Writer(), map[string]any{
			"str":   "foo",
			"num":   100,
			"bool":  false,
			"null":  nil,
			"array": []string{"foo", "bar", "baz"},
			"obj":   map[string]any{"a": 1, "b": 2},
		})
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	//
	// Run regular mode
	//
	topicName := flag.Arg(0)
	profilePath := flag.Arg(1)

	if topicName == "" || profilePath == "" {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "mwtail: [options] <topic> <filename>")
		fmt.Fprintln(os.Stderr, "\truns a console consumer on the given topic using the provided connection profile")

		if *helpFlag {
			os.Exit(0)
		}

		os.Exit(64) // EX_USAGE
	}

	// Load the Mister Webhooks connection profile from a file on the filesystem.
	profile, err := client.LoadConnectionProfile(profilePath)

	if err != nil {
		log.Fatal(err)
	}

	// Create a consumer that reads generic events from `topicName`.
	consumer, err := client.NewConsumer(
		profile,
		client.DeclareWebhookTopic[any](topicName),
	)

	if err != nil {
		log.Fatal(err)
	}

	// Loop endlessly (or until a network error) reading input events and dumping them to stdout.
	err = consumer.Consume(context.Background(), func(ctx context.Context, event *client.Webhook[any]) error {
		return render(log.Writer(), event)
	})

	if err != nil {
		log.Fatalf("error: %s", err)
	}
}
