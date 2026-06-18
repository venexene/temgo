package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/venexene/temgo/internal/timer"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("Welcome to Temgo!")

	prolog := flag.Duration("p", 10 * time.Second, "prologue time")
	work := flag.Duration("w", 25 * time.Minute, "work time")
	rest := flag.Duration("r",  5 * time.Minute, "rest time")
	longRest := flag.Duration("lr",  30 * time.Minute, "long rest time")
	cycles := flag.Int("c",  4, "cycles num")
	sprints := flag.Int("s",  3, "sprints num")

	flag.Parse()

	timer := timer.NewWorkTimer(*prolog, *work, *rest, *longRest, *cycles, *sprints)
	if err := timer.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
	}
	
	fmt.Println("Bye!")
}