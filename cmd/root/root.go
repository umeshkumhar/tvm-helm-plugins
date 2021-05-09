package root

import (
	"log"

	"github.com/spf13/cobra"
	pre_upgrade "github.com/trilioData/tvm-helm-plugins/pkg/pre-upgrade"
)

const (
	binaryName               = "operator-init"
	helmReleaseFlag          = "release"
	helmReleaseNamespaceFlag = "namespace"
	serviceAccountFlag       = "serviceaccount"
	upgradeHookUsage         = "Pre upgrade job of the TVM v2.0.x helm releases to the new TVM v2.1.x release"

	shortUsage = "upgrade-plugin is used to run pre upgrade job from v2.0.x to v2.1.x TVM operator"
	longUsage  = `upgrade-plugin is used to run the pre upgrade job before upgrade from v2.0.x helm release to the new 
v2.1.x release version of k8s-triliovault-operator.

--release         <release_name> of the previous k8s-triliovault-operator
--namespace       <namespace> of the previous k8s-triliovault-operator 
--service-account <service_account> name which user wants to run job with. It will take default service account`
)

var (
	rootCmd              *cobra.Command
	helmReleaseName      string
	helmReleaseNamespace string
	serviceAccount       string
)

func init() {
	rootCmd = newHelmUpgradeCmd()
}

func newHelmUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   binaryName,
		Short: shortUsage,
		Long:  longUsage,
		RunE:  runHelmPreUpgradeJobE,
	}

	cmd.Flags().StringVarP(&helmReleaseName, helmReleaseFlag, "r", "", upgradeHookUsage)
	cmd.Flags().StringVarP(&helmReleaseNamespace, helmReleaseNamespaceFlag, "n", "", upgradeHookUsage)
	cmd.Flags().StringVarP(&serviceAccount, serviceAccountFlag, "s", "default", upgradeHookUsage)
	err := cmd.MarkFlagRequired(helmReleaseFlag)
	if err != nil {
		log.Fatal("Error while setting up the Hook command")
	}

	cmd.SetHelpFunc(rootHelp)

	return cmd
}

func runHelmPreUpgradeJobE(cmd *cobra.Command, args []string) error {
	if pre_upgrade.Validate(helmReleaseName, helmReleaseNamespace, serviceAccount) {
		if err := pre_upgrade.Do(helmReleaseName, helmReleaseNamespace, serviceAccount); err != nil {
			return err
		}
	}

	return nil
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}

func rootHelp(cmd *cobra.Command, args []string) {
	cmd.Println()
	cmd.Println(longUsage)
	cmd.Println()
	rootUsage(cmd)
}

// the usage subset of help info, results the list of actions available for the binary
func rootUsage(cmd *cobra.Command) {
	actions := cmd.Commands()
	cmd.Println("Usage:")
	cmd.Printf("  %s [action] [flags]\n", binaryName)
	cmd.Println()
	cmd.Println("possible actions:")
	for _, a := range actions {
		cmd.Printf("  %s\n", a.Use)
	}
	cmd.Println()
	cmd.Println("For more help, run hook --help")
}
