package utils

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/gin-gonic/gin"
)

func PrintAppBanner() {
	// Define styles for the information box
	infoBoxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")). // Cyan text
		Border(lipgloss.RoundedBorder()).
		Padding(0, 10).
		Align(lipgloss.Center)

	// Information to display
	info := fmt.Sprintf(`
InscructHub API
Version: %s
Gin Version: %s
IP: http://127.0.0.1:%s

Mongodb: Successfully connected
`, os.Getenv("VERSION"), gin.Version, os.Getenv("PORT"))

	// Print the information box after server starts
	fmt.Println(infoBoxStyle.Render(info))
}
