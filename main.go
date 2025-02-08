package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/bogem/id3v2/v2"
)

var (
	debugFlag  = flag.Bool("debug", false, "debug mode")
	dryFlag    = flag.Bool("dry", false, "dry run")
	targetFlag = flag.String("target", "", "target file")

	artistFlag = flag.String("artist", "", "artist name")
	albumFlag  = flag.String("album", "", "album name")
)

func isFlagSet(name string) bool {
	var found bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("Error", "error", err)
		os.Exit(1)
	}
	slog.Info("Done")
}

func run(ctx context.Context) error {
	flag.Parse()
	setupLogger()
	if *dryFlag {
		slog.InfoContext(ctx, "Dry run is enabled")
	}
	if *targetFlag == "" {
		return fmt.Errorf("target is required")
	}

	slog.Info("Input", "target", *targetFlag)
	files, err := filepath.Glob(*targetFlag)
	if err != nil {
		return fmt.Errorf("list files: %w", err)
	}

	for _, file := range files {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := processFile(file); err != nil {
			return fmt.Errorf("process file: %w", err)
		}
	}
	return nil
}

func processFile(target string) error {
	tag, err := id3v2.Open(target, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("open file tag: %w", err)
	}

	slog.Info("Opened",
		"file", target,
		"id3v2-version", tag.Version(),
	)

	slog.Info("Current tag",
		"artist", tag.Artist(),
		"album", tag.Album(),
	)

	if isFlagSet("artist") {
		slog.Debug("Set artist")
		tag.SetArtist(*artistFlag)
	}
	if isFlagSet("album") {
		slog.Debug("Set album")
		tag.SetAlbum(*albumFlag)
	}

	slog.Info("Updated tag",
		"artist", tag.Artist(),
		"album", tag.Album(),
	)

	if *dryFlag {
		return nil
	}

	if err := tag.Save(); err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	return nil
}

func setupLogger() {
	level := slog.LevelInfo
	if *debugFlag {
		level = slog.LevelDebug
	}

	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(h)
	slog.SetDefault(logger)
}
