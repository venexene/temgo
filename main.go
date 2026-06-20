package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/venexene/temgo/internal/config"
	"github.com/venexene/temgo/internal/timer"
)


func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("Welcome to Temgo!")

	params, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
		os.Exit(2)
	}

	wt := timer.NewWorkTimer(params.Prolog, params.Work, params.Rest, params.LongRest, params.Cycles, params.Sprints)
	
	if err := wt.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
	}
	
	fmt.Println("Bye!")
}