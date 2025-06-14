package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/meldiron/odyc-cli/cmd"
)

func main() {
	// Override log styles
	styles := log.DefaultStyles()

	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR!").
		Bold(true).
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("1")).
		Foreground(lipgloss.Color("0"))

	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString("DEBUG").
		Bold(true).
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("8")).
		Foreground(lipgloss.Color("0"))

	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("INFO").
		Bold(true).
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("6")).
		Foreground(lipgloss.Color("0"))

	styles.Levels[3] = lipgloss.NewStyle().
		SetString("ODYC").
		Bold(true).
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("165")).
		Foreground(lipgloss.Color("0"))

	styles.Levels[2] = lipgloss.NewStyle().
		SetString("SUCCESS").
		Bold(true).
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("40")).
		Foreground(lipgloss.Color("0"))

	log.SetStyles(styles)
	log.SetReportTimestamp(false)
	log.SetLevel(-4)

	cmd.Execute()
}
