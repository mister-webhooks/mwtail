// Package main contains an example client for Mister Webhooks.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"time"

	"github.com/mister-webhooks/mister-webhooks-client/golang/client"
	"github.com/nickwells/col.mod/v5/col"
	"github.com/nickwells/col.mod/v5/colfmt"
)

const MODE_TOPIC = "topic"

func main() {
	mode := MODE_TOPIC

	helpFlag := flag.Bool("help", false, "show this usage message")

	flag.Func("mode", "set the operating mode", func(s string) error {
		switch s {
		case MODE_TOPIC:
			mode = s
			return nil
		default:
			return fmt.Errorf("%s is not a valid mode, valid modes are {%s}", s, "foo")
		}
	})

	//
	// Run regular mode
	//
	flag.Parse()
	profilePath := flag.Arg(0)

	if profilePath == "" {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "mwtop: [options] <filename>")
		fmt.Fprintln(os.Stderr, "\truns a monitor provided connection profile")

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

	admin, err := client.NewAdmin(context.Background(), profile)

	if err != nil {
		log.Fatal(err)
	}

	oldTopicStats := []client.Topic[client.PartitionStat]{}
	lastChange := time.Now()

	for {
		topicStats, err := admin.StatTopics(context.Background())
		newTime := time.Now()

		if err != nil {
			log.Fatal(err)
		}

		if !reflect.DeepEqual(oldTopicStats, topicStats) {
			lastChange = newTime
			oldTopicStats = topicStats
		}

		fmt.Print("\033[H\033[2J")
		fmt.Printf("%s\n", newTime.Format(time.ANSIC))
		fmt.Printf("  last change: %s ago\n", newTime.Sub(lastChange).Round(time.Second))
		fmt.Printf("\n")

		switch mode {
		case MODE_TOPIC:
			if len(topicStats) == 0 {
				continue
			}

			maxPartitionWidth := math.Ceil(math.Log10(float64(maxOffset(topicStats))))

			partitionColumns := make([]*col.Col, maxPartitions(topicStats))

			for i := range maxPartitions(topicStats) {
				partitionColumns[i] = col.New(&colfmt.Int{W: uint(maxPartitionWidth)}, fmt.Sprintf("[%d]", i))
			}

			report, err := col.NewReport(
				col.NewHeaderOrPanic(col.HdrOptDontUnderline),
				os.Stdout,
				col.New(&colfmt.WrappedString{W: uint(maxTopicLength(topicStats))}, "topic"),
				partitionColumns...,
			)

			if err != nil {
				log.Fatal(err)
			}

			for _, topicStat := range topicStats {
				lastOffsets := make([]int64, len(topicStat.PartitionData))

				for i, partitionData := range topicStat.PartitionData {
					lastOffsets[i] = partitionData.LastOffset
				}

				columns := make([]any, 0, len(lastOffsets)+1)
				columns = append(columns, topicStat.Name)

				for _, offset := range lastOffsets {
					columns = append(columns, offset)
				}

				report.PrintRow(columns...)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func maxTopicLength[T any](topicStats []client.Topic[T]) int {
	maxLen := 0

	for _, topicStat := range topicStats {
		if l := len(topicStat.Name); l > maxLen {
			maxLen = l
		}
	}

	return maxLen
}

func maxPartitions[T any](topicStats []client.Topic[T]) int {
	maxPartitions := 0

	for _, topicStat := range topicStats {
		if p := len(topicStat.PartitionData); p > maxPartitions {
			maxPartitions = p
		}
	}

	return maxPartitions
}

func maxOffset(topicStats []client.Topic[client.PartitionStat]) int64 {
	maxOffset := int64(0)

	for _, topicStat := range topicStats {
		for _, partitionData := range topicStat.PartitionData {
			if maxOffset > partitionData.LastOffset {
				maxOffset = partitionData.LastOffset
			}
		}
	}

	return maxOffset
}
