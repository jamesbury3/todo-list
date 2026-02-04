package model

import "github.com/charmbracelet/lipgloss"

// Color styles
var (
	// View tabs
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117")).Background(lipgloss.Color("236")).Padding(0, 1)
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)

	// Cursor and items
	cursorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	todoTextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	timestampStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	updateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Italic(true)

	// Headers and sections
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true)
	countStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)

	// Input prompts
	promptStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	inputCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	helpTextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)

	// Messages
	successMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	errorMessageStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	infoMessageStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))

	// Commands
	commandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
)
