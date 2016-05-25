package main

import (
//"zdd/lib/config"
//"fmt"
	"flag"
	"os"
	"fmt"
	"zdd/lib/command/deploy"
	"zdd/lib/config"
	"zdd/lib/command/rollback"
)

func main() {
	// define command flag sets

	// deploy
	deployCmd := flag.NewFlagSet("deploy", flag.ExitOnError)
	deployVersionFlag := deployCmd.String("v", "", "version to deploy")
	deployConfigFlag := deployCmd.String("c", "", "configuration file")

	// rollback
	rollbackCmd := flag.NewFlagSet("rollback", flag.ExitOnError)
	rollbackConfigFile := rollbackCmd.String("c", "", "configuration file")

	if len(os.Args) == 1 {
		// display help
		fmt.Println("Usage")
		os.Exit(0)
	}

	// choose which command to use
	switch os.Args[1] {
	case "deploy":
		deployCmd.Parse(os.Args[2:])
		break

	case "rollback":
		rollbackCmd.Parse(os.Args[2:])
		break

	case "help":
		// display help about command
		break

	default:
		fmt.Printf("Unknown: %s\n", os.Args[1])
		os.Exit(config.COMMANDLINE_ERROR)
		break
	}

	// deploy command
	if deployCmd.Parsed() {
		if *deployConfigFlag == "" {
			fmt.Println("Please provide config file")
			os.Exit(config.COMMANDLINE_ERROR)
		}

		if *deployVersionFlag == "" {
			fmt.Println("Please provide version")
			os.Exit(config.COMMANDLINE_ERROR)
		}

		// run deploy command
		yml := config.ReadFile(*deployConfigFlag)
		cfg := config.ParseConfig(yml, config.GetEnvMap())
		deploy.Run(cfg, *deployVersionFlag)
	}

	// rollback command
	if rollbackCmd.Parsed() {
		if *rollbackConfigFile == "" {
			fmt.Println("Please provide config file")
			os.Exit(config.COMMANDLINE_ERROR)
		}

		// run rollback command
		yml := config.ReadFile(*rollbackConfigFile)
		cfg := config.ParseConfig(yml, config.GetEnvMap())
		rollback.Run(cfg)
	}
}