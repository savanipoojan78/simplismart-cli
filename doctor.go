package main
import (
	// "context"
	"fmt"
	"os"
	"os/exec"

	// "k8s.io/client-go/tools/clientcmd"
	"github.com/spf13/cobra"
	// "k8s.io/client-go/tools/clientcmd/api"
)

// New doctor command to check if Helm is installed
var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check if Helm is installed",
	Run: func(cmd *cobra.Command, args []string) {
		// Logic to check if Helm is installed
		if _, err := exec.LookPath("helm"); err != nil {
			fmt.Println("Helm is not installed.")
			os.Exit(1)
		}
		fmt.Println("Helm is installed.")
		versionCmd := exec.Command("helm", "version", "--short")
		versionOutput, err := versionCmd.Output()
		if err != nil {
			fmt.Println("Error retrieving Helm version:", err)
			os.Exit(1)
		}
		fmt.Printf("Helm version: %s\n", string(versionOutput))
	},
}