package azurelogs

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/wtfutil/wtf/modules/azurelogs/session"
)

// logQueryClient abstracts the Azure logs client so it can be mocked in tests.
type logQueryClient interface {
	QueryWorkspace(ctx context.Context, workspaceID string, body azquery.Body, options *azquery.LogsClientQueryWorkspaceOptions) (azquery.LogsClientQueryWorkspaceResponse, error)
}

type TableRow []string
type TableResp struct {
	Header []string
	Rows   []TableRow
}

func RunQuery(sess *session.Session, client logQueryClient, cf session.QueryFile) (*TableResp, error) {
	sess.Logger.Info("config file", "cf", cf)

	var err error
	var tableResp TableResp
	tableResp.Header = cf.Columns

	if cf.WorkspaceID == "" {
		return nil, fmt.Errorf("azure workspace ID is required but not configured")
	}

	if cf.SubscriptionID == "" {
		return nil, fmt.Errorf("azure subscription ID is required but not configured")
	}

	if client == nil {
		client, err = session.GetLogsClient(sess, cf.SubscriptionID)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure Logs client for subscription %s: %w", cf.SubscriptionID, err)
		}
	}

	res, err := client.QueryWorkspace(
		context.Background(),
		cf.WorkspaceID,
		azquery.Body{
			Query: to.Ptr(cf.Query),
		},
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query on workspace %s: %w", cf.WorkspaceID, err)
	}

	if res.Error != nil {
		return nil, res.Error
	}

	switch len(res.Tables) {
	case 0:
		return nil, fmt.Errorf("query returned no data tables: %s", cf.Query)
	case 1:
		if len(res.Tables[0].Columns) == 0 {
			return nil, fmt.Errorf("query returned table with no columns: %s", cf.Query)
		}
	default:
		return nil, fmt.Errorf("query returned %d tables, expected 1: %s", len(res.Tables), cf.Query)
	}

	for _, row := range res.Tables[0].Rows {
		var r TableRow

		for f := range row {
			if row[f] == nil {
				continue
			}

			switch t := row[f].(type) {
			case string:
				r = append(r, t)
			case float64:
				r = append(r, fmt.Sprintf("%.0f", t))
			}
		}
		tableResp.Rows = append(tableResp.Rows, r)
	}

	return &tableResp, nil
}
