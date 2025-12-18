package printutils

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
)

// PrintHealthDetails prints only the detailed health check results (no summary)
func PrintHealthDetails(noHeaders bool, results []data.HealthCheckResult) {
	if len(results) == 0 {
		fmt.Println("All resources are healthy!")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !noHeaders {
		fmt.Fprintln(w, "AWS PROFILE\tAWS REGION\tCLUSTER NAME\tKIND\tNAMESPACE\tNAME\tREADY\tSTATUS\tMESSAGE")
	}

	for _, r := range results {
		namespace := r.Namespace
		if namespace == "" {
			namespace = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.Profile,
			r.Region,
			r.ClusterName,
			r.Kind,
			namespace,
			r.Name,
			r.Ready,
			r.Status,
			r.Message,
		)
	}

	w.Flush()
}

// PrintHealthSummary prints only the cluster health summary (no details)
func PrintHealthSummary(noHeaders bool, summaries []data.ClusterHealthSummary) {
	if len(summaries) == 0 {
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !noHeaders {
		fmt.Fprintln(w, "AWS PROFILE\tAWS REGION\tCLUSTER NAME\tPODS\tDEPLOYMENTS\tSTATEFULSETS\tDAEMONSETS\tREPLICASETS\tSTATUS")
	}

	for _, s := range summaries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d/%d\t%d/%d\t%d/%d\t%d/%d\t%d/%d\t%s\n",
			s.Profile,
			s.Region,
			s.ClusterName,
			s.HealthyPods, s.TotalPods,
			s.HealthyDeployments, s.TotalDeployments,
			s.HealthyStatefulSets, s.TotalStatefulSets,
			s.HealthyDaemonSets, s.TotalDaemonSets,
			s.HealthyReplicaSets, s.TotalReplicaSets,
			s.OverallStatus,
		)
	}

	w.Flush()
}
