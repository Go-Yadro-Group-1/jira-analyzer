package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	analyzerv1 "github.com/Go-Yadro-Group-1/Jira-Analyzer/gen/grpc/analyzer/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultServerAddr = "localhost:50051"

var errUnknownChartType = errors.New("unknown chart type")

//nolint:exhaustruct
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "gRPC client for manual testing",
	}

	cmd.PersistentFlags().String("addr", defaultServerAddr, "gRPC server address")

	cmd.AddCommand(newGetStatsCommand())
	cmd.AddCommand(newGetChartCommand())
	cmd.AddCommand(newCompareCommand())

	return cmd
}

func dial(cmd *cobra.Command) (*grpc.ClientConn, error) {
	addr, err := cmd.InheritedFlags().GetString("addr")
	if err != nil {
		return nil, fmt.Errorf("get addr flag: %w", err)
	}

	//nolint:exhaustruct
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	return conn, nil
}

// get-stats

//nolint:exhaustruct
func newGetStatsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "get-stats",
		Short:        "Get statistics for a project",
		Example:      "  jira-analyzer client get-stats --project-id 1",
		RunE:         runGetStats,
		SilenceUsage: true,
	}

	cmd.Flags().Int32("project-id", 0, "project ID (required)")
	_ = cmd.MarkFlagRequired("project-id")

	return cmd
}

func runGetStats(cmd *cobra.Command, _ []string) error {
	projectID, err := cmd.Flags().GetInt32("project-id")
	if err != nil {
		return fmt.Errorf("get project-id flag: %w", err)
	}

	conn, err := dial(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := analyzerv1.NewAnalyzerServiceClient(conn).GetStats(
		context.Background(),
		&analyzerv1.GetStatsRequest{ProjectId: projectID},
	)
	if err != nil {
		return fmt.Errorf("GetStats: %w", err)
	}

	return printJSON(resp.GetStats())
}

// get-chart

//nolint:exhaustruct
func newGetChartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-chart",
		Short: "Get chart data for a project",
		Example: "  jira-analyzer client get-chart --project-id 1 --chart-type open_state_histogram\n" +
			"  available types: open_state_histogram, state_distribution, complexity_histogram, priority, daily_activity",
		RunE:         runGetChart,
		SilenceUsage: true,
	}

	cmd.Flags().Int32("project-id", 0, "project ID (required)")
	_ = cmd.MarkFlagRequired("project-id")

	cmd.Flags().String("chart-type", "", "chart type (required)")
	_ = cmd.MarkFlagRequired("chart-type")

	return cmd
}

func parseChartType(name string) (analyzerv1.ChartType, error) {
	chartTypeNames := map[string]analyzerv1.ChartType{
		"open_state_histogram": analyzerv1.ChartType_CHART_TYPE_OPEN_STATE_HISTOGRAM,
		"state_distribution":   analyzerv1.ChartType_CHART_TYPE_STATE_DISTRIBUTION,
		"complexity_histogram": analyzerv1.ChartType_CHART_TYPE_COMPLEXITY_HISTOGRAM,
		"priority":             analyzerv1.ChartType_CHART_TYPE_PRIORITY,
		"daily_activity":       analyzerv1.ChartType_CHART_TYPE_DAILY_ACTIVITY,
	}

	chartType, ok := chartTypeNames[name]
	if !ok {
		return 0, fmt.Errorf(
			"%w %q, available: open_state_histogram, state_distribution, complexity_histogram, priority, daily_activity",
			errUnknownChartType,
			name,
		)
	}

	return chartType, nil
}

func runGetChart(cmd *cobra.Command, _ []string) error {
	projectID, err := cmd.Flags().GetInt32("project-id")
	if err != nil {
		return fmt.Errorf("get project-id flag: %w", err)
	}

	chartTypeName, err := cmd.Flags().GetString("chart-type")
	if err != nil {
		return fmt.Errorf("get chart-type flag: %w", err)
	}

	chartType, err := parseChartType(chartTypeName)
	if err != nil {
		return err
	}

	conn, err := dial(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := analyzerv1.NewAnalyzerServiceClient(conn).GetChart(
		context.Background(),
		&analyzerv1.GetChartRequest{ProjectId: projectID, ChartType: chartType},
	)
	if err != nil {
		return fmt.Errorf("GetChart: %w", err)
	}

	// data is already JSON — pretty-print it
	var raw json.RawMessage = resp.GetData()

	return printJSON(raw)
}

// compare

//nolint:exhaustruct
func newCompareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "compare",
		Short:        "Compare statistics of two projects",
		Example:      "  jira-analyzer client compare --project-id-a 1 --project-id-b 2",
		RunE:         runCompare,
		SilenceUsage: true,
	}

	cmd.Flags().Int32("project-id-a", 0, "first project ID (required)")
	_ = cmd.MarkFlagRequired("project-id-a")

	cmd.Flags().Int32("project-id-b", 0, "second project ID (required)")
	_ = cmd.MarkFlagRequired("project-id-b")

	return cmd
}

func runCompare(cmd *cobra.Command, _ []string) error {
	projectIDA, err := cmd.Flags().GetInt32("project-id-a")
	if err != nil {
		return fmt.Errorf("get project-id-a flag: %w", err)
	}

	projectIDB, err := cmd.Flags().GetInt32("project-id-b")
	if err != nil {
		return fmt.Errorf("get project-id-b flag: %w", err)
	}

	conn, err := dial(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := analyzerv1.NewAnalyzerServiceClient(conn).CompareProjects(
		context.Background(),
		&analyzerv1.CompareProjectsRequest{ProjectIdA: projectIDA, ProjectIdB: projectIDB},
	)
	if err != nil {
		return fmt.Errorf("CompareProjects: %w", err)
	}

	return printJSON(resp)
}

func printJSON(value any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	err := enc.Encode(value)
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}

	return nil
}
