// Package ui handles the TUI (Text User Interface) for displaying
// BTRFS snapshots. It uses the 'bubbles' library to create a list-based UI
// where snapshots are displayed with details like ID, description, method, and date.
package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	reflow "github.com/muesli/reflow/truncate"
	"github.com/xeyossr/btrfs-rollback/internal/btrfs"
)

// snapshotItem wraps around Snapshot to make it compatible with the list UI
type snapshotItem struct {
	btrfs.Snapshot
}

// Title returns the ID of the snapshot to display as the title in the list
func (s snapshotItem) Title() string { return s.ID }

// FilterValue returns the ID, used for searching/filtering in the list
func (s snapshotItem) FilterValue() string { return s.ID }

// Description combines Method and Date for displaying them in the list
func (s snapshotItem) Description() string {
	return fmt.Sprintf("%s | %s", s.Method, s.Date)
}

// UI styles for various components
var (
	titleStyle    = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("205")).Bold(true)
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("236")).
			Bold(true)
	appStyle = lipgloss.NewStyle().Margin(1, 2)
)

// snapshotDelegate defines how the snapshot items are rendered in the list
type snapshotDelegate struct {
	width int // Width of the list (for dynamic layout)
}

// Height returns the height of each list item (just one line per item)
func (d snapshotDelegate) Height() int { return 1 }

// Spacing returns the spacing between items (no spacing)
func (d snapshotDelegate) Spacing() int { return 0 }

// Update is called to handle updates (currently not needed)
func (d snapshotDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

// Render formats and prints the snapshot item in the list
func (d snapshotDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	// Cast item to snapshotItem
	snapshotItem, ok := item.(snapshotItem)
	if !ok {
		return
	}

	// Dynamic width calculations for each column in the list
	maxWidth := d.width - 6 // Reserve space for the prefix
	idWidth := maxWidth / 20
	descriptionWidth := (maxWidth * 2) / 3
	methodWidth := maxWidth / 10
	dateWidth := maxWidth / 4

	// If Description is empty, use "N/A" as the default value
	descriptionStr := snapshotItem.DescriptionStr
	if descriptionStr == "" {
		descriptionStr = "N/A" // Or just use empty space
	}

	// Use reflow to adjust text length
	id := reflow.StringWithTail(snapshotItem.ID, uint(idWidth), "...")
	description := reflow.StringWithTail(descriptionStr, uint(descriptionWidth), "...")
	method := reflow.StringWithTail(snapshotItem.Method, uint(methodWidth), "…")
	date := reflow.StringWithTail(snapshotItem.Date, uint(dateWidth), "…")

	// Format the line: each column has a dynamic width
	line := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s", idWidth, id, descriptionWidth, description, methodWidth, method, dateWidth, date)

	// Prefix for selected item
	prefix := "  "
	if index == m.Index() {
		prefix = "> "
		line = selectedStyle.Render(line)
	}

	// Print the final output
	fmt.Fprint(w, prefix+line)
}

// model struct holds the state of the TUI UI, including the list of snapshots and delegate
type model struct {
	width  int
	height int

	quitting bool
	err      error

	list       list.Model       // The list of snapshot items
	selectedID string           // The ID of the selected snapshot
	snapshots  []btrfs.Snapshot // List of snapshots to display
	delegate   snapshotDelegate // Delegate responsible for rendering
}

// Init initializes the program and prepares it for the TUI
func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

// Update handles events such as window resizing and user input
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-4)

		m.delegate.width = msg.Width - 6
		m.list.SetDelegate(m.delegate)

		// Update the list items based on the snapshots
		items := make([]list.Item, len(m.snapshots))
		for i, s := range m.snapshots {
			items[i] = snapshotItem{Snapshot: s}
		}
		m.list.SetItems(items)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(m.snapshots) > 0 {
				m.selectedID = m.snapshots[m.list.Index()].ID
				return m, tea.Quit
			}
		case "ctrl+c", "esc", "q":
			m.quitting = true
			m.err = errors.New("")
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the current view of the snapshots in the list
func (m *model) View() string {
	if len(m.snapshots) == 0 {
		return appStyle.Render("❌ No snapshots found")
	}
	return appStyle.Render(m.list.View())
}

// InitialModel creates and initializes the TUI model with snapshot data
func InitialModel(snapshots []btrfs.Snapshot) *model {
	items := make([]list.Item, len(snapshots))
	for i, s := range snapshots {
		items[i] = snapshotItem{Snapshot: s}
	}

	delegate := snapshotDelegate{width: 80} // Initial width

	// Create a new list with the delegate and items
	l := list.New(items, delegate, 0, 0)
	l.Title = "BTRFS Snapshots"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = list.Styles{}.PaginationStyle
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	return &model{
		list:      l,
		snapshots: snapshots,
		delegate:  delegate,
	}
}

// RunUI runs the TUI program and returns the selected snapshot ID
func RunUI(snapshots []btrfs.Snapshot) (string, error) {
	p := tea.NewProgram(InitialModel(snapshots), tea.WithAltScreen())

	m, err := p.Run()
	if err != nil {
		return "", err
	}

	model := m.(*model)
	if model.err != nil {
		os.Exit(1)
	}
	return model.selectedID, nil
}

// Confirm displays a confirmation prompt and waits for a Y/n response.
// If user presses 'Y' or 'y', returns true immediately (without enter).
// If 'N', empty, or any other key, returns false.
// Ctrl+C exits the program immediately.
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)

	// Set terminal to raw mode to capture single key without enter
	oldState, err := term.MakeRaw(uintptr(syscall.Stdin))
	if err != nil {
		fmt.Println("\nUnable to capture input.")
		return false
	}
	defer term.Restore(uintptr(syscall.Stdin), oldState)

	var b = make([]byte, 1)
	os.Stdin.Read(b)

	fmt.Println(string(b)) // echo the pressed key

	switch strings.ToLower(string(b)) {
	case "y":
		return true
	default:
		return false
	}
}

// Clear screen
func ClearScreen() {
	exec.Command("clear").Run()
}
