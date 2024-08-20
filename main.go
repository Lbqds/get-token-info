package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	sdk "github.com/alephium/go-sdk"
)

type Client struct {
	timeout time.Duration
	impl    *sdk.APIClient
}

func NewClient(endpoint string, apiKey string, timeout int) *Client {
	configuration := sdk.NewConfiguration()
	var host string
	if strings.HasPrefix(endpoint, "http://") {
		host = endpoint[7:]
	}
	if strings.HasPrefix(endpoint, "https://") {
		host = endpoint[8:]
	}
	configuration.Host = host

	if apiKey != "" {
		configuration.AddDefaultHeader("X-API-KEY", apiKey)
	}
	return &Client{
		timeout: time.Duration(timeout) * time.Second,
		impl:    sdk.NewAPIClient(configuration),
	}
}

func (c *Client) timeoutContext(ctx context.Context) (*time.Time, context.Context, context.CancelFunc) {
	timestamp := time.Now()
	timeoutCtx, cancel := context.WithDeadline(ctx, timestamp.Add(c.timeout))
	return &timestamp, timeoutCtx, cancel
}

func (c *Client) MultiCallContract(ctx context.Context, multiCall *sdk.MultipleCallContract) (*sdk.MultipleCallContractResult, error) {
	_, timeoutCtx, cancel := c.timeoutContext(ctx)
	defer cancel()

	request := c.impl.ContractsApi.PostContractsMulticallContract(timeoutCtx).MultipleCallContract(*multiCall)
	response, _, err := request.Execute()
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) GetTokenInfo(ctx context.Context, contractAddress string, groupIndex int) (*sdk.MultipleCallContractResult, error) {
	multiCallContract := &sdk.MultipleCallContract{
		Calls: []sdk.CallContract{
			{
				Group:       int32(groupIndex),
				Address:     contractAddress,
				MethodIndex: 0,
			},
			{
				Group:       int32(groupIndex),
				Address:     contractAddress,
				MethodIndex: 1,
			},
			{
				Group:       int32(groupIndex),
				Address:     contractAddress,
				MethodIndex: 2,
			},
		},
	}
	return c.MultiCallContract(ctx, multiCallContract)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify the node url")
		return
	}

	apiKey := ""
	if len(os.Args) == 3 {
		apiKey = os.Args[2]
	}
	client := NewClient(os.Args[1], apiKey, 10)
	ctx := context.Background()
	result, err := client.GetTokenInfo(ctx, "27HxXZJBTPjhHXwoF1Ue8sLMcSxYdxefoN2U6d8TKmZsm", 0)
	if err != nil {
		fmt.Printf("Request error: %v\n", err)
		return
	}
	object, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return
	}
	fmt.Println(string(object))
}
