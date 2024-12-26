package cmd

import (
	"os"
	"strings"

	"github.com/chrishrb/go-grip/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-grip [file]",
	Short: "Render markdown document as html",
	Args:  cobra.MatchAll(cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		theme, _ := cmd.Flags().GetString("theme")
		browser, _ := cmd.Flags().GetBool("browser")
		openReadme, _ := cmd.Flags().GetBool("readme")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		boundingBox, _ := cmd.Flags().GetBool("bounding-box")

		client := pkg.Client{
			Theme:       strings.ToLower(theme),
			OpenBrowser: browser,
			Host:        host,
			Port:        port,
			OpenReadme:  openReadme,
			BoundingBox: boundingBox,
		}

		var file string
		if len(args) == 1 {
			file = args[0]
		}
		err := client.Serve(file)
		cobra.CheckErr(err)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().String("theme", "auto", "Select css theme [light/dark/auto]")
	rootCmd.Flags().BoolP("browser", "b", true, "Open new browser tab")
	rootCmd.Flags().BoolP("readme", "r", true, "Open readme if no file provided")
	rootCmd.Flags().StringP("host", "H", "localhost", "Host to use")
	rootCmd.Flags().IntP("port", "p", 6419, "Port to use")
	rootCmd.Flags().BoolP("bounding-box", "B", true, "Add bounding box to HTML")
}
