package middleware

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/gin-gonic/gin"
)

// Make the logger better looks
func CustomLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// start time
		startTime := time.Now()

		// process request
		c.Next()

		// calculate execution time
		latency := time.Since(startTime)

		// request status code
		statusCode := c.Writer.Status()

		// Define styles
		dateStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")) // Blue text

		statusStyles := map[int]lipgloss.Style{
			2: lipgloss.NewStyle().Foreground(lipgloss.Color("2")), // Green
			3: lipgloss.NewStyle().Foreground(lipgloss.Color("3")), // Yellow
			4: lipgloss.NewStyle().Foreground(lipgloss.Color("5")), // Magenta
			5: lipgloss.NewStyle().Foreground(lipgloss.Color("1")), // Red
		}

		methodStyles := map[string]lipgloss.Style{
			"GET":    lipgloss.NewStyle().Foreground(lipgloss.Color("2")), // Green
			"POST":   lipgloss.NewStyle().Foreground(lipgloss.Color("3")), // Yellow
			"PUT":    lipgloss.NewStyle().Foreground(lipgloss.Color("4")), // Blue
			"DELETE": lipgloss.NewStyle().Foreground(lipgloss.Color("1")), // Red
		}

		// Select styles based on status code and method
		statusStyle, ok := statusStyles[statusCode/100]
		if !ok {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // Default
		}

		methodStyle, exists := methodStyles[c.Request.Method]
		if !exists {
			methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // Default
		}

		// custom log format and color
		fmt.Printf("%s | %s | %s | %v | %s | %-7s %s\n",
			dateStyle.Render(time.Now().Format("2006/01/02 - 15:04:05")),
			statusStyle.Render(fmt.Sprintf("%d", statusCode)),
			methodStyle.Render(c.Request.Method),
			latency,
			c.ClientIP(),
			c.Request.URL.Path,
			lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(""), // request path color
		)
	}
}