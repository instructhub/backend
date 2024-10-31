package utils

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/gin-gonic/gin"
)

// For print app banner when server start
func PrintAppBanner() {
	// Define styles for the information box
	infoBoxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")). // Cyan text
		Border(lipgloss.RoundedBorder()).
		Padding(0, 10).
		Align(lipgloss.Center)

	// Information to display
	info := fmt.Sprintf(`
InstructHub API
Version: %s
Gin Version: %s
Domain: %s

`, os.Getenv("VERSION"), gin.Version, os.Getenv("BASE_URL"))

	// Print the information box after server starts
	fmt.Println(infoBoxStyle.Render(info))
}
