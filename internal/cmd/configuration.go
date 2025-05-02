package cmd

import (
    "fmt"
    "strings"
    
    "github.com/ayushsarode/orb/internal/config"
    "github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config [key] [value]",
        Short: "Get and set repository or global options",
        Args:  cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.LoadConfig()
            if err != nil {
                return fmt.Errorf("loading config: %w", err)
            }
            
            key := args[0]
            
            // Get configuration value
            if len(args) == 1 {
                value := cfg.Get(key)
                if value == "" {
                    return fmt.Errorf("no value set for %s", key)
                }
                fmt.Println(value)
                return nil
            }
            
            // Set configuration value
            value := strings.Join(args[1:], " ")
            cfg.Set(key, value)
            
            if err := cfg.Save(); err != nil {
                return fmt.Errorf("saving config: %w", err)
            }
            
            return nil
        },
    }
    
    return cmd
}