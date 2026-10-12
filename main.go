package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/blastjax/golang-redis/internal/api"
	"github.com/blastjax/golang-redis/internal/config"
	"github.com/blastjax/golang-redis/internal/database"
	"github.com/blastjax/golang-redis/internal/github"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "sync-users":
		syncUsersCommand(cfg)
	case "list-users":
		listUsersCommand(cfg)
	case "setup-db":
		setupDBCommand(cfg)
	case "serve":
		serveCommand(cfg)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// newFlagSet creates the flags for a subcommand. Its -h output is generated from
// the flag definitions, so it can't drift from them.
func newFlagSet(name, description string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s %s [options]\n\n%s\n\n", os.Args[0], name, description)
		fs.PrintDefaults()
	}
	return fs
}

func syncUsersCommand(cfg *config.Config) {
	fs := newFlagSet("sync-users", "Fetches users from the GitHub Users REST API, optionally storing them in PostgreSQL")
	count := fs.Int("count", 30, "Number of users to fetch")
	since := fs.Int("since", 0, "Fetch users with an ID greater than this (0 with -store: continue after the last stored user)")
	format := fs.String("format", "table", "Output format: table, json, detailed")
	store := fs.Bool("store", false, "Store users in the database")
	fs.Parse(os.Args[2:])

	// Initialize database if storing users
	var userRepo *database.UserRepository
	if *store {
		db, err := database.New(&cfg.Database)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		userRepo = database.NewUserRepository(db)

		// Auto-detect since value if not provided and storing
		if *since == 0 {
			lastID, err := userRepo.GetLastSyncedUserID()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting last synced user ID: %v\n", err)
				os.Exit(1)
			}
			*since = lastID
		}
	}

	startTime := time.Now()

	users, err := github.NewClient(&cfg.GitHub).FetchAllUsers(context.Background(), *count, *since)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching users: %v\n", err)
		os.Exit(1)
	}

	// Status messages go to stderr (log) so -format json output stays valid JSON
	log.Printf("Fetched %d users in %v", len(users), time.Since(startTime))

	// Store users in database if requested
	if *store && len(users) > 0 {
		if err := userRepo.InsertUsers(users); err != nil {
			fmt.Fprintf(os.Stderr, "Error storing users: %v\n", err)
			os.Exit(1)
		}
		log.Printf("Stored %d users in database", len(users))
	}

	// Display users
	if err := displayUsers(users, *format); err != nil {
		fmt.Fprintf(os.Stderr, "Error displaying users: %v\n", err)
		os.Exit(1)
	}
}

func listUsersCommand(cfg *config.Config) {
	fs := newFlagSet("list-users", "Lists users from the PostgreSQL database")
	limit := fs.Int("limit", 30, "Number of users to list")
	offset := fs.Int("offset", 0, "Offset for pagination")
	format := fs.String("format", "table", "Output format: table, json, detailed")
	fs.Parse(os.Args[2:])

	// Connect to database
	db, err := database.New(&cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	userRepo := database.NewUserRepository(db)

	// Get users from database
	users, err := userRepo.GetUsers(*limit, *offset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching users from database: %v\n", err)
		os.Exit(1)
	}

	// Get total count
	total, err := userRepo.GetTotalUsers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting total user count: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Showing %d of %d total users (offset: %d)", len(users), total, *offset)

	// Display users
	if err := displayUsers(users, *format); err != nil {
		fmt.Fprintf(os.Stderr, "Error displaying users: %v\n", err)
		os.Exit(1)
	}
}

func setupDBCommand(cfg *config.Config) {
	newFlagSet("setup-db", "Initializes the PostgreSQL database schema for storing GitHub users").Parse(os.Args[2:])

	// Connect to database
	db, err := database.New(&cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create schema
	if err := db.CreateSchema(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating database schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Database schema created successfully!")
}

func serveCommand(cfg *config.Config) {
	newFlagSet("serve", "Starts the REST API server (endpoints and settings are listed in README.md)").Parse(os.Args[2:])

	// Create and start the API server
	server, err := api.NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating API server: %v\n", err)
		os.Exit(1)
	}
	defer server.Close()

	// Start the server (this blocks until shutdown)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting API server: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `GitHub Users CLI Tool

Usage: %[1]s <command> [options]

Commands:
  sync-users    Fetch users from GitHub API
  list-users    List users from database
  setup-db      Initialize database schema
  serve         Start REST API server
  help          Show this help message

Use '%[1]s <command> -h' for more information about a command.
`, os.Args[0])
}

func displayUsers(users []github.User, format string) error {
	switch format {
	case "json":
		return github.DisplayUsersJSON(users)
	case "table":
		return github.DisplayUsersTable(users)
	case "detailed":
		return github.DisplayUsersDetailed(users)
	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, detailed)", format)
	}
}
