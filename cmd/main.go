package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/dhananjayjamwal78/taskmgr/internal/store"
    "github.com/dhananjayjamwal78/taskmgr/internal/ui"
    
)

func main() {
    // 1) Init store
    s, err := store.NewBoltStore("tasks.db")
    if err != nil {
        log.Fatalf("store init: %v", err)
    }
    defer s.Close()

    // 2) Check subcommand
    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "add":
            cmdAdd(s, os.Args[2:])
        case "list":
            cmdList(s, os.Args[2:])
        case "done":
            cmdDone(s, os.Args[2:])
        case "rm":
            cmdRm(s, os.Args[2:])
        default:
            fmt.Printf("Unknown command %q\n", os.Args[1])
            usage()
            os.Exit(1)
        }
        return
    }

    // 3) No args → launch TUI (we’ll build this later)
    launchTUI(s)
}

// usage prints help.
func usage() {
    fmt.Println(`Usage:
  taskmgr add <text> --due YYYY-MM-DD    Add a new task
  taskmgr list [--all]                  List tasks (default: pending)
  taskmgr done <id>                     Mark task done
  taskmgr rm <id>                       Remove task
  taskmgr                               Launch interactive UI`)
}

// cmdAdd parses flags and calls AddTask.
func cmdAdd(s *store.Store, args []string) {
    fs := flag.NewFlagSet("add", flag.ExitOnError)
    dueStr := fs.String("due", "", "due date YYYY-MM-DD")
    fs.Parse(args)

    text := fs.Arg(0)
    if text == "" || *dueStr == "" {
        fmt.Println("Usage: taskmgr add <text> --due YYYY-MM-DD")
        os.Exit(1)
    }
    due, err := time.Parse("2006-01-02", *dueStr)
    if err != nil {
        log.Fatalf("invalid due date: %v", err)
    }

    id, err := s.AddTask(text, due)
    if err != nil {
        log.Fatalf("add task: %v", err)
    }
    fmt.Printf("Added task #%d\n", id)
}

// cmdList shows tasks.
func cmdList(s *store.Store, args []string) {
    fs := flag.NewFlagSet("list", flag.ExitOnError)
    all := fs.Bool("all", false, "include completed tasks")
    fs.Parse(args)

    tasks, err := s.ListTasks(*all)
    if err != nil {
        log.Fatalf("list tasks: %v", err)
    }
    if len(tasks) == 0 {
        fmt.Println("No tasks found.")
        return
    }
    for _, t := range tasks {
        status := " "
        if t.Done {
            status = "✓"
        }
        fmt.Printf("[%s] #%d: %s (due %s)\n",
            status, t.ID, t.Text, t.Due.Format("2006-01-02"))
    }
}

// cmdDone marks a task done.
func cmdDone(s *store.Store, args []string) {
    if len(args) != 1 {
        fmt.Println("Usage: taskmgr done <id>")
        os.Exit(1)
    }
    var id uint64
    fmt.Sscanf(args[0], "%d", &id)
    if err := s.MarkDone(id); err != nil {
        log.Fatalf("mark done: %v", err)
    }
    fmt.Printf("Marked #%d done\n", id)
}

// cmdRm deletes a task.
func cmdRm(s *store.Store, args []string) {
    if len(args) != 1 {
        fmt.Println("Usage: taskmgr rm <id>")
        os.Exit(1)
    }
    var id uint64
    fmt.Sscanf(args[0], "%d", &id)
    if err := s.DeleteTask(id); err != nil {
        log.Fatalf("delete task: %v", err)
    }
    fmt.Printf("Removed #%d\n", id)
}

// launchTUI is a placeholder for our future TUI.
func launchTUI(s *store.Store) {
	
    fmt.Println("🚧 TUI not implemented yet – coming soon!")
    ui := ui.New(s)
   if err := ui.Run(); err != nil {
   log.Fatalf("Error running UI: %v", err)
  }
}
