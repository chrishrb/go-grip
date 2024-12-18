//go:build debug
// +build debug

package cmd

import (
	"errors"

	"github.com/chrishrb/go-grip/internal"
	"github.com/spf13/cobra"
)

var emojiscraperCmd = &cobra.Command{
	Use:   "emojiscraper [emoji-out] [emoji-map-out]",
	Short: "Scrape emojis from gist",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cobra.CheckErr(errors.New("provide exact 2 arguments"))
		}
		internal.ScrapeEmojis(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(emojiscraperCmd)
}
