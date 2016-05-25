package rollback

import (
	"fmt"
	"os"

	"zdd/lib/config"
	"zdd/lib/release"
)

// perform deploy command
func Run(cfg config.Config) {
	build, err := release.CreateRollback(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := build.LoadImage(); err != nil {
		fmt.Println("Failed to locate image..")
		os.Exit(1)
	}

	if err := build.CreateContainer(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := build.StartContainer(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := build.ExecCommands(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := build.CheckAlive(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := build.UpdateUpstream(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := build.StopOld(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := build.Finalize(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}
