/*
  MiningHQ Miner Manager - The MiningHQ Miner Manager GUI
  https://mininghq.io

	Copyright (C) 2018  Donovan Solms     <https://github.com/donovansolms>
                                        <https://github.com/mininghq>

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/donovansolms/mininghq-miner-manager/src/installer"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// AppName is injected by the Astilectron bundler
var AppName string

// Asset is injected by the Astilectron bundler
var Asset bootstrap.Asset

// RestoreAssets is injected by the Astilectron bundler
var RestoreAssets bootstrap.RestoreAssets

// installCmd represents a fresh installation
var installCmd = &cobra.Command{
	Use:   "MinerManager",
	Short: "The MiningHQ Miner Manager",
	Long: `
   __  ____      _           __ ______
  /  |/  (_)__  (_)__  ___ _/ // / __ \
 / /|_/ / / _ \/ / _ \/ _ '/ _  / /_/ /
/_/  /_/_/_//_/_/_//_/\_, /_//_/\___\_\
                    /___/ Miner Manager


The MiningHQ Manager installs and configures the MiningHQ
services required for managing your rigs. It can be run as a GUI or command line.

Once installed, the Miner Manager can be used to view basic miner
stats on the local machine, however, the MiningHQ Dashboard
(https://www.mininghq.io/dashboard) is the best place to monitor miners from.`,
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, err := homedir.Dir()
		if err != nil {
			fmt.Printf("Unable to get user home directory: %s\n", err)
		}

		mhqInstaller, err := installer.New(homeDir, runtime.GOOS, apiEndpoint)
		if err != nil {
			fmt.Printf("Unable to create installer: %s\n", err)
			return
		}

		// We're not checking for noGUI here since --uninstall is a command-line
		// only operation for now
		if mustUninstall {

			// Get the current installed path
			installedCheckfilePath := filepath.Join(homeDir, ".mhqpath")
			installedPath, err := ioutil.ReadFile(installedCheckfilePath)
			if err != nil {
				fmt.Printf(`
We were unable to find the installed location for the MiningHQ services. Please
remove the files manually where you installed the services.
				`)
				fmt.Println()
				os.Exit(0)
			}

			err = mhqInstaller.UninstallSync(strings.TrimSpace(string(installedPath)), installedCheckfilePath)
			if err != nil {
				fmt.Println("ERR", err)
				return
			}

			os.Exit(0)
		}

		if noGUI {
			err = mhqInstaller.InstallSync()
			if err != nil {
				fmt.Println("ERR", err)
			}
			return
		}

		// If the '--no-gui' flag wasn't specified, we'll start the Electron
		// interface
		// AppName, Asset and RestoreAssets are injected by the bundler
		gui, err := installer.NewGUI(
			AppName,
			Asset,
			RestoreAssets,
			homeDir,
			runtime.GOOS,
			apiEndpoint,
			debug,
		)
		if err != nil {
			// Setting the output to stdout so the user can see the error
			log.SetOutput(os.Stdout)
			log.Fatalf("Unable to set up miner: %s", err)
		}

		err = gui.Run()
		if err != nil {
			// Setting the output to stdout so the user can see the error
			log.SetOutput(os.Stdout)
			log.Fatalf("Unable to run miner: %s", err)
		}

	},
}

func init() {
	installCmd.Flags().BoolVar(&debug, "debug", false, "Run the manager in debug mode, a log file will be created")
	installCmd.Flags().BoolVar(&noGUI, "no-gui", false, "Run the manager without GUI")
	installCmd.Flags().StringVar(&apiEndpoint, "api-endpoint", "http://mininghq.local/api/v1", "The base API endpoint for MiningHQ")
	installCmd.Flags().BoolVar(&mustUninstall, "uninstall", false, "Completely remove MiningHQ services from this system")
}
