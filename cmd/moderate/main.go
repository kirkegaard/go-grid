package main

import (
	"flag"
	"fmt"
	"github.com/kirkegaard/go-grid/internal/server"

	"os"
)

func main() {
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	kickCmd := flag.NewFlagSet("kick", flag.ExitOnError)

	// Define flags for the 'kick' command
	kickClientID := kickCmd.String("clientid", "", "The ID of the client to kick")

	// Check which subcommand is invoked
	if len(os.Args) < 2 {
		fmt.Println("Expected 'list' or 'kick' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		// Parse the list subcommand
		listCmd.Parse(os.Args[2:])
		clients := server.GetHub().GetConnectedClients()
		for _, client := range clients {
			fmt.Println(client)
		}

	case "kick":
		// Parse the kick subcommand
		kickCmd.Parse(os.Args[2:])

		// Validate the required flags for the 'kick' command
		if *kickClientID == "" {
			fmt.Println("Please provide a client ID using --clientid")
			kickCmd.PrintDefaults()
			os.Exit(1)
		}

		// Execute the kick logic
		if server.GetHub().KickClient(*kickClientID) {
			fmt.Printf("Client %s kicked successfully.\n", *kickClientID)
		} else {
			fmt.Printf("Client %s not found.\n", *kickClientID)
		}

	default:
		fmt.Println("Unknown command. Expected 'list' or 'kick'")
		os.Exit(1)
	}

}
