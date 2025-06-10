package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
)

var (
	log *logrus.Logger
	ds  *YAMLDatastore
)

const asciiArt = `
          _____
         /\    \
        /::\____\
       /DFML|   |        
      /Will:|   |        
     /Gordon|   |        
    /:::/|::|   |        
   /GNU/ |::|   |        
  /GPL/  |::|___|______
 /2.0/   |Project:\    \ 
/:::/    |Maradona:\____\
\::/    / ~~~~~/:::/ D  /
 \/____/      /:::/ E  / 
             /:::/ E  /
            /:::/ P  /   
           /:::/ F  /    
          /:::/ I  /     
         /:::/ E  /      
        /:::/ L  /       
        \::/ D  /        
         \/____/
`

func init() {
	// Initialize logger
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logFile, err := os.OpenFile("yarndb.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	} else {
		log.Warn("Failed to open log file, using stderr")
	}
	logLevel := viper.GetString("log_level")
	if level, err := logrus.ParseLevel(logLevel); err == nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	// Initialize viper for configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("data_dir", "data")
	viper.SetDefault("auto_save_interval", 60)
	viper.SetDefault("log_level", "info")
	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("No config file found, using defaults: %v", err)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "yarndb",
		Short: "YarnDB: A thread-safe, high-speed in-memory database for YAML data",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(asciiArt)
			fmt.Println("YarnDB: Use --help for available commands")
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize datastore
			dataDir := viper.GetString("data_dir")
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				log.Fatalf("Failed to create data directory: %v", err)
			}
			var err error
			ds, err = NewYAMLDatastore(dataDir)
			if err != nil {
				log.Fatalf("Failed to initialize YarnDB: %v", err)
			}
			// Ensure final save on exit
			defer func() {
				log.Info("YarnDB exiting, saving final state...")
				if err := ds.Save(); err != nil {
					log.Errorf("Final save error: %v", err)
				}
			}()
		},
	}

	rootCmd.PersistentFlags().String("data-dir", viper.GetString("data_dir"), "Directory for YAML files")
	rootCmd.PersistentFlags().Int("auto-save-interval", viper.GetInt("auto_save_interval"), "Auto-save interval in seconds")
	rootCmd.PersistentFlags().String("log-level", viper.GetString("log_level"), "Log level (debug, info, warn, error)")
	viper.BindPFlag("data_dir", rootCmd.PersistentFlags().Lookup("data-dir"))
	viper.BindPFlag("auto_save_interval", rootCmd.PersistentFlags().Lookup("auto-save-interval"))
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	// Init command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize YarnDB",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(asciiArt)
			log.Info("YarnDB initialized successfully")
			fmt.Println("YarnDB initialized in", ds.dir)
		},
	})

	// Set command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "set <recordID> <yaml>",
		Short: "Create or update a record with YAML data",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			recordID, yamlStr := args[0], args[1]
			if !isValidID(recordID) {
				log.Error("Invalid recordID: must be alphanumeric")
				fmt.Println("Error: recordID must be alphanumeric")
				return
			}
			var data interface{}
			if err := yaml.Unmarshal([]byte(yamlStr), &data); err != nil {
				log.Errorf("Invalid YAML: %v", err)
				fmt.Printf("Error: invalid YAML: %v\n", err)
				return
			}
			fileID := strings.Split(recordID, "_")[0] // Simple fileID derivation
			err := ds.Set(recordID, data, fileID)
			if err != nil {
				log.Errorf("Failed to set record: %v", err)
				fmt.Printf("Error: %v\n", err)
				return
			}
			log.Infof("Set record %s in file %s", recordID, fileID)
			fmt.Printf("Set record %s\n", recordID)
		},
	})

	// Get command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "get <recordID>",
		Short: "Retrieve a record by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			recordID := args[0]
			if !isValidID(recordID) {
				log.Error("Invalid recordID: must be alphanumeric")
				fmt.Println("Error: recordID must be alphanumeric")
				return
			}
			data, err := ds.Get(recordID)
			if err != nil {
				log.Errorf("Failed to get record: %v", err)
				fmt.Printf("Error: %v\n", err)
				return
			}
			if data == nil {
				log.Warnf("Record %s not found", recordID)
				fmt.Printf("Record %s not found\n", recordID)
				return
			}
			out, err := yaml.Marshal(data)
			if err != nil {
				log.Errorf("Failed to marshal record: %v", err)
				fmt.Printf("Error: %v\n", err)
				return
			}
			log.Infof("Retrieved record %s", recordID)
			fmt.Printf("Record %s:\n%s\n", recordID, string(out))
		},
	})

	// Delete command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "delete <recordID>",
		Short: "Delete a record by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			recordID := args[0]
			if !isValidID(recordID) {
				log.Error("Invalid recordID: must be alphanumeric")
				fmt.Println("Error: recordID must be alphanumeric")
				return
			}
			err := ds.Delete(recordID)
			if err != nil {
				log.Errorf("Failed to delete record: %v", err)
				fmt.Printf("Error: %v\n", err)
				return
			}
			log.Infof("Deleted record %s", recordID)
			fmt.Printf("Deleted record %s\n", recordID)
		},
	})

	// Query command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "query <expression>",
		Short: "Query records with a key=value expression (e.g., department=engineering)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			expr := args[0]
			parts := strings.SplitN(expr, "=", 2)
			if len(parts) != 2 {
				log.Error("Invalid query expression: must be key=value")
				fmt.Println("Error: query expression must be key=value")
				return
			}
			key, value := parts[0], parts[1]
			if !isValidKey(key) {
				log.Error("Invalid key: must be alphanumeric with dots")
				fmt.Println("Error: key must be alphanumeric with dots")
				return
			}
			records, err := ds.Query(key, value)
			if err != nil {
				log.Errorf("Query failed: %v", err)
				fmt.Printf("Error: %v\n", err)
				return
			}
			log.Infof("Queried %d records for %s=%s", len(records), key, value)
			fmt.Printf("Found %d records:\n", len(records))
			for id, data := range records {
				out, _ := yaml.Marshal(data)
				fmt.Printf("- %s:\n%s\n", id, string(out))
			}
		},
	})

	// Index command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "index <key>",
		Short: "Create an index on a top-level key for faster queries",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			if !isValidKey(key) {
				log.Error("Invalid key: must be alphanumeric with dots")
				fmt.Println("Error: key must be alphanumeric with dots")
				return
			}
			err := ds.CreateIndex(key)
			if err != nil {
				log.Errorf("Failed to create index: %v", err)
				fmt.Printf("Error: %v\n", err)
				return
			}
			log.Infof("Created index on key %s", key)
			fmt.Printf("Index created on %s\n", key)
		},
	})

	// Transaction command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "trans",
		Short: "Start an interactive transaction",
		Run: func(cmd *cobra.Command, args []string) {
			tx, err := ds.BeginTransaction()
			if err != nil {
				log.Errorf("Failed to start transaction: %v", err)
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("Transaction started. Commands: set, delete, commit, rollback")
			for {
				fmt.Print("tx> ")
				var input string
				fmt.Scanln(&input)
				parts := strings.Fields(input)
				if len(parts) == 0 {
					continue
				}
				switch parts[0] {
				case "set":
					if len(parts) != 3 {
						fmt.Println("Usage: set <recordID> <yaml>")
						continue
					}
					var data interface{}
					if err := yaml.Unmarshal([]byte(parts[2]), &data); err != nil {
						fmt.Printf("Error: invalid YAML: %v\n", err)
						continue
					}
					fileID := strings.Split(parts[1], "_")[0]
					if err := tx.Set(parts[1], data, fileID); err != nil {
						fmt.Printf("Error: %v\n", err)
						continue
					}
					fmt.Printf("Set record %s\n", parts[1])
				case "delete":
					if len(parts) != 2 {
						fmt.Println("Usage: delete <recordID>")
						continue
					}
					if err := tx.Delete(parts[1]); err != nil {
						fmt.Printf("Error: %v\n", err)
						continue
					}
					fmt.Printf("Deleted record %s\n", parts[1])
				case "commit":
					if err := tx.Commit(); err != nil {
						fmt.Printf("Error: %v\n", err)
						continue
					}
					log.Info("Transaction committed")
					fmt.Println("Transaction committed")
					return
				case "rollback":
					tx.Rollback()
					log.Info("Transaction rolled back")
					fmt.Println("Transaction rolled back")
					return
				default:
					fmt.Println("Unknown command. Use: set, delete, commit, rollback")
				}
			}
		},
	})

	// Save command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "save",
		Short: "Manually save YarnDB",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ds.Save(); err != nil {
				log.Errorf("Save failed: %v", err)
				fmt.Printf("Error: %v\n", err)
				return
			}
			log.Info("YarnDB saved manually")
			fmt.Println("YarnDB saved")
		},
	})

	// Status command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show YarnDB status",
		Run: func(cmd *cobra.Command, args []string) {
			ds.mu.RLock()
			recordCount := len(ds.data)
			fileCount := len(ds.files)
			dirty := ds.dirty
			indexes := make([]string, 0, len(ds.indexes))
			for key := range ds.indexes {
				indexes = append(indexes, key)
			}
			ds.mu.RUnlock()
			log.Info("Displayed YarnDB status")
			fmt.Printf("YarnDB Status:\n")
			fmt.Printf("- Records: %d\n", recordCount)
			fmt.Printf("- Files: %d\n", fileCount)
			fmt.Printf("- Indexes: %v\n", indexes)
			fmt.Printf("- Dirty (unsaved changes): %v\n", dirty)
			fmt.Printf("- Data directory: %s\n", ds.dir)
			fmt.Printf("- Auto-save interval: %ds\n", viper.GetInt("auto_save_interval"))
		},
	})

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("YarnDB command execution failed: %v", err)
		os.Exit(1)
	}
}

// isValidID checks if an ID is alphanumeric
func isValidID(id string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(id)
}

// isValidKey checks if a key is alphanumeric with dots
func isValidKey(key string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9.]+$`).MatchString(key)
}
