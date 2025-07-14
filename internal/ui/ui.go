package ui

import (
    "fmt"
    "time"

    "github.com/rivo/tview"
    "github.com/gdamore/tcell/v2"

    "github.com/dhananjayjamwal78/taskmgr/internal/store"
	"github.com/dhananjayjamwal78/taskmgr/pkg/task"

)

// UI holds the tview application and our data store.
type UI struct {
    app    *tview.Application
    pages  *tview.Pages
    list   *tview.List
    form   *tview.Form
    status *tview.TextView
    store  *store.Store
	tasks  []task.Task
}

// New creates the UI for the given store.
func New(s *store.Store) *UI {
    ui := &UI{
        app:   tview.NewApplication(),
        pages: tview.NewPages(),
        store: s,
    }

    // Status bar at bottom
    ui.status = tview.NewTextView().
        SetTextAlign(tview.AlignCenter).
        SetText("Press 'a' to Add, 'q' to Quit")


    ui.buildListView()
    ui.buildFormView()

    flex := tview.NewFlex().SetDirection(tview.FlexRow).
        AddItem(ui.pages, 0, 1, true).
        AddItem(ui.status, 1, 0, false)

    ui.app.SetRoot(flex, true)
    return ui
}


func (ui *UI) buildListView() {
    ui.list = tview.NewList().
        ShowSecondaryText(false).
        SetDoneFunc(func() { ui.app.Stop() })

    ui.pages.AddPage("list", ui.list, true, true)

    // Global keybinding: 'a' → switch to form
    ui.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Rune() {
        case 'a':
            ui.pages.SwitchToPage("form")
            return nil
        case 'q':
            ui.app.Stop()
            return nil
			  case 'd': 
            idx := ui.list.GetCurrentItem()
            if idx < len(ui.tasks) {
                id := ui.tasks[idx].ID
                if err := ui.store.MarkDone(id); err == nil {
                    ui.flashStatus(fmt.Sprintf("Marked #%d done", id))
                    ui.refreshList()
                }
            }
            return nil
        case 'x': 
            idx := ui.list.GetCurrentItem()
            if idx < len(ui.tasks) {
                id := ui.tasks[idx].ID
                if err := ui.store.DeleteTask(id); err == nil {
                    ui.flashStatus(fmt.Sprintf("Deleted #%d", id))
                    ui.refreshList()
                }
            }
         return nil
        }
        return event
    })
}

// buildFormView sets up the add-task form.
func (ui *UI) buildFormView() {
    ui.form = tview.NewForm().
        AddInputField("Task", "", 0, nil, nil).
        AddInputField("Due (YYYY-MM-DD)", "", 0, nil, nil).
        AddButton("Save", ui.onFormSave).
        AddButton("Cancel", func() { ui.pages.SwitchToPage("list") }).
        SetButtonsAlign(tview.AlignCenter)

    ui.pages.AddPage("form", ui.form, true, false)
}

// onFormSave is called when user presses Save on the form.
func (ui *UI) onFormSave() {
    taskText := ui.form.GetFormItemByLabel("Task").(*tview.InputField).GetText()
    dueStr   := ui.form.GetFormItemByLabel("Due (YYYY-MM-DD)").(*tview.InputField).GetText()

    due, err := time.Parse("2006-01-02", dueStr)
    if err != nil {
        ui.flashStatus(fmt.Sprintf("Invalid date: %v", err))
        return
    }
    if taskText == "" {
        ui.flashStatus("Task description cannot be empty")
        return
    }

    if _, err := ui.store.AddTask(taskText, due); err != nil {
        ui.flashStatus(fmt.Sprintf("Error saving: %v", err))
        return
    }
    ui.flashStatus("Task added!")
    ui.form.Clear(true)
    ui.refreshList()
    ui.pages.SwitchToPage("list")
}

func (ui *UI) refreshList() {
    ui.list.Clear()

    // Fetch only pending tasks
    ts, _ := ui.store.ListTasks(false)
    ui.tasks = ts

    // If no tasks, show placeholder
    if len(ts) == 0 {
        ui.list.AddItem("(no pending tasks)", "", 0, nil)
        return
    }

    now := time.Now()
    for _, t := range ts {
        // Base label
        base := fmt.Sprintf("#%d %s (due %s)",
            t.ID, t.Text, t.Due.Format("2006-01-02"))

        // Wrap in color tags
        var label string
        switch {
        case t.Due.Before(now):
            // Overdue → red
            label = "[red]" + base + "[-]"
        case t.Due.Before(now.Add(24 * time.Hour)):
            // Due within 24h → yellow
            label = "[yellow]" + base + "[-]"
        default:
            // No color
            label = base
        }

        ui.list.AddItem(label, "", 0, nil)
    }
}

// flashStatus shows a temporary message in the status bar.
func (ui *UI) flashStatus(msg string) {
    ui.status.SetText(msg)
    go func() {
        time.Sleep(2 * time.Second)
        ui.app.QueueUpdateDraw(func() {
            ui.status.SetText("Press 'a' to Add, 'q' to Quit")
        })
    }()
}

// Run starts the application.
func (ui *UI) Run() error {
    ui.refreshList()
    return ui.app.Run()
}
