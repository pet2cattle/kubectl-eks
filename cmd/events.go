package cmd

import (
	"context"
	"log"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Show Kubernetes events across namespaces",
	Long: `Display Kubernetes events across all or specific namespaces, sorted by timestamp.

Events provide insights into cluster activities such as pod scheduling,
image pulls, volume mounts, configuration changes, and errors.

By default shows all event types (Normal and Warning). Use --warnings-only
to filter for just Warning events. Events are sorted with most recent first.`,
	Example: `  # Show all events across all namespaces
  kubectl eks events

  # Show only warning events
  kubectl eks events --warnings-only

  # Show events for specific namespace
  kubectl eks events -n kube-system

  # Show all events (explicit)
  kubectl eks events --all`,
	Run: func(cmd *cobra.Command, args []string) {
		clusterInfo, err := GetCurrentClusterInfo()
		if err != nil {
			log.Fatalf("Error getting current cluster info: %v", err)
		}

		namespace, _ := cmd.Flags().GetString("namespace")
		allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")
		warningsOnly, _ := cmd.Flags().GetBool("warnings-only")
		allEvents, _ := cmd.Flags().GetBool("all")
		noHeaders, _ := cmd.Flags().GetBool("no-headers")

		// Default to all namespaces unless specific namespace is provided
		if !allNamespaces && namespace == "" {
			allNamespaces = true
		}

		if allNamespaces {
			namespace = ""
		}

		// Show all events by default (both Normal and Warning)
		if !warningsOnly && !allEvents {
			allEvents = true
		}

		// If --warnings-only is explicitly set, only show warnings
		if warningsOnly {
			allEvents = false
		}

		events, err := k8s.GetEvents(context.Background(), namespace)
		if err != nil {
			log.Fatalf("Error getting events: %v", err)
		}

		if len(events) == 0 {
			if namespace == "" {
				log.Println("No events found in any namespace")
			} else {
				log.Printf("No events found in namespace: %s\n", namespace)
			}
			return
		}

		eventInfos := make([]data.EventInfo, 0)
		for _, event := range events {
			// Filter by type if warnings-only is set
			if warningsOnly && !strings.EqualFold(event.Type, "Warning") {
				continue
			}

			// Use EventTime if LastTimestamp is not set
			lastSeen := event.LastTimestamp.Time
			if lastSeen.IsZero() && !event.EventTime.Time.IsZero() {
				lastSeen = event.EventTime.Time
			}

			info := data.EventInfo{
				Profile:     clusterInfo.AWSProfile,
				Region:      clusterInfo.Region,
				ClusterName: clusterInfo.ClusterName,
				Namespace:   event.Namespace,
				LastSeen:    lastSeen,
				Type:        event.Type,
				Reason:      event.Reason,
				Object:      event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
				Message:     event.Message,
				Count:       event.Count,
			}
			eventInfos = append(eventInfos, info)
		}

		if len(eventInfos) == 0 {
			if warningsOnly {
				log.Println("No warning events found. Try --all to see all event types.")
			} else {
				log.Println("No events match the specified criteria")
			}
			return
		}

		printutils.PrintEvents(noHeaders, eventInfos...)
	},
}

func init() {
	eventsCmd.Flags().StringP("namespace", "n", "", "Namespace to show events for")
	eventsCmd.Flags().BoolP("all-namespaces", "A", false, "Show events across all namespaces (default)")
	eventsCmd.Flags().Bool("warnings-only", false, "Show only warning events")
	eventsCmd.Flags().Bool("all", false, "Show all events (default behavior)")
	rootCmd.AddCommand(eventsCmd)
}
