package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/ramiayoub/logbook/backend/internal/store"
)

// The server binary. Running it with no subcommand serves the API; the other
// subcommands are the account operations that app/docs/security.md says must
// exist on the server and not over HTTP.
//
// Account management is here rather than in an endpoint on purpose: there is no
// self-service registration route to defend because there is no registration
// route at all.
const usage = `logbook-server -- the logbook API

usage:
  logbook-server [flags]                serve the API
  logbook-server createuser <name>      create an account (prompts for the password)
  logbook-server passwd <name>          change a password and revoke every session
  logbook-server users                  list accounts
  logbook-server disable <name>         lock an account out immediately
  logbook-server enable <name>          unlock an account

flags and their environment variables:
  -db     LOGBOOK_DB      path to the SQLite file      (default /var/lib/logbook/logbook.db)
  -addr   LOGBOOK_ADDR    listen address               (default 127.0.0.1:8081)
  -origin LOGBOOK_ORIGIN  the site's scheme and host   (default https://ayoub.fi)
  -holder LOGBOOK_HOLDER  licence holder's name, printed on the exported PDFs
  -insecure-cookie        drop the Secure flag; local plain-HTTP development only
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("logbook-server", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	var (
		dbPath   = fs.String("db", env("LOGBOOK_DB", "/var/lib/logbook/logbook.db"), "SQLite file")
		addr     = fs.String("addr", env("LOGBOOK_ADDR", "127.0.0.1:8081"), "listen address")
		origin   = fs.String("origin", env("LOGBOOK_ORIGIN", "https://ayoub.fi"), "site origin")
		holder   = fs.String("holder", env("LOGBOOK_HOLDER", ""), "licence holder's name for the PDF exports")
		insecure = fs.Bool("insecure-cookie", false, "drop Secure on the session cookie")
	)

	positional, err := parseFlagsAnywhere(fs, args)
	if err != nil {
		return err
	}

	// Anything positional is a subcommand and its arguments.
	if len(positional) > 0 {
		db, err := store.Open(*dbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		return runSubcommand(positional[0], positional[1:], db)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	db, err := store.Open(*dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Refuse to serve an empty database rather than presenting an empty
	// logbook as though it were the truth. Rule 0.2: never silently show a
	// wrong record.
	n, err := db.CountFlights()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%s holds no flights; run `logbookctl import` before serving", *dbPath)
	}

	if *insecure && strings.HasPrefix(*origin, "https://") {
		return errors.New("-insecure-cookie with an https origin: that combination is only ever a mistake")
	}

	srv := NewServer(db, Config{
		Addr:         *addr,
		Origin:       *origin,
		HolderName:   *holder,
		SecureCookie: !*insecure,
		Logger:       logger,
	})

	httpSrv := &http.Server{
		Addr:    *addr,
		Handler: srv,
		// Bounded on every axis. This process shares a 1 vCPU / 2 GB box with
		// the owner's other sites and must not be the thing that exhausts it
		// (rule 0.3).
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    16 << 10,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go srv.housekeeping(ctx)

	errc := make(chan error, 1)
	go func() {
		logger.Info("serving", "addr", *addr, "db", *dbPath, "flights", n, "origin", *origin)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
		}
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpSrv.Shutdown(shutdownCtx)
	}
}

// parseFlagsAnywhere parses flags wherever they appear and returns the
// positional arguments in order.
//
// Go's flag package stops at the first non-flag argument, so with a plain
// fs.Parse the -db in `createuser rami -db /tmp/x.db` is never read -- the
// command would silently reach for the production database instead. That is
// the wrong direction to be wrong in on a legal record, and it is the order an
// operator actually types. Parsing in a loop, taking one positional at a time,
// accepts every arrangement.
func parseFlagsAnywhere(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	rest := args
	for {
		if err := fs.Parse(rest); err != nil {
			return nil, err
		}
		rest = fs.Args()
		if len(rest) == 0 {
			return positional, nil
		}
		positional = append(positional, rest[0])
		rest = rest[1:]
	}
}

// housekeeping sweeps expired sessions and stale rate-limiter entries. Both are
// self-healing without it -- LookupSession drops an expired row on sight and
// Allow forgets an idle key -- so this only bounds the tables when nobody is
// asking.
func (s *Server) housekeeping(ctx context.Context) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if n, err := s.db.PurgeExpiredSessions(); err != nil {
				s.log.Error("purging sessions", "error", err)
			} else if n > 0 {
				s.log.Info("purged expired sessions", "count", n)
			}
			s.limiter.Prune()
		}
	}
}

func runSubcommand(cmd string, args []string, db *store.DB) error {
	needName := func() (string, error) {
		if len(args) != 1 || args[0] == "" {
			return "", fmt.Errorf("usage: logbook-server %s <username>", cmd)
		}
		return args[0], nil
	}

	switch cmd {
	case "createuser":
		name, err := needName()
		if err != nil {
			return err
		}
		pw, err := promptNewPassword()
		if err != nil {
			return err
		}
		u, err := db.CreateUser(name, pw)
		if err != nil {
			return err
		}
		fmt.Printf("created user %q (id %d)\n", u.Username, u.ID)
		return nil

	case "passwd":
		name, err := needName()
		if err != nil {
			return err
		}
		pw, err := promptNewPassword()
		if err != nil {
			return err
		}
		if err := db.SetPassword(name, pw); err != nil {
			return err
		}
		fmt.Printf("password changed for %q; every session was revoked\n", name)
		return nil

	case "users":
		list, err := db.Users()
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("no users; create one with `logbook-server createuser <name>`")
			return nil
		}
		for _, u := range list {
			state := "active"
			if u.Disabled {
				state = "DISABLED"
			}
			sessions, err := db.Sessions(u.ID)
			if err != nil {
				return err
			}
			fmt.Printf("%-20s %-8s created %s  %d session(s)\n",
				u.Username, state, u.CreatedAt.Format("2006-01-02"), len(sessions))
		}
		return nil

	case "disable", "enable":
		name, err := needName()
		if err != nil {
			return err
		}
		disable := cmd == "disable"
		if err := db.SetUserDisabled(name, disable); err != nil {
			return err
		}
		if disable {
			// Locking someone out must take effect now. Their live sessions
			// already stop working, but removing the rows makes that visible
			// rather than implicit.
			n, err := revokeAllFor(db, name)
			if err != nil {
				return err
			}
			fmt.Printf("disabled %q and revoked %d session(s)\n", name, n)
			return nil
		}
		fmt.Printf("enabled %q\n", name)
		return nil

	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown subcommand %q", cmd)
	}
}

func revokeAllFor(db *store.DB, username string) (int, error) {
	list, err := db.Users()
	if err != nil {
		return 0, err
	}
	for _, u := range list {
		if u.Username == username {
			return db.RevokeAllSessions(u.ID)
		}
	}
	return 0, fmt.Errorf("no user %q", username)
}

// promptNewPassword reads a password twice without echoing it.
//
// Read from the terminal rather than taken as an argument: a command-line
// argument lands in the shell history and in the process list, where every
// other user of this shared box can read it.
func promptNewPassword() (string, error) {
	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		return "", errors.New("a password must be typed at a terminal, not piped: " +
			"piping it would put it in the shell history and the process list")
	}
	fmt.Print("password: ")
	first, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", err
	}
	fmt.Print("again: ")
	second, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", err
	}
	if string(first) != string(second) {
		return "", errors.New("the two passwords do not match")
	}
	return string(first), nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
