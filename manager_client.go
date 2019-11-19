package main

import (
	"context"

	pb "github.com/cytobot/rpc/manager"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type ManagerClient struct {
	client pb.ManagerClient
}

func NewManagerClient(managerAddress string) (*ManagerClient, error) {
	conn, err := grpc.Dial(managerAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	//defer conn.Close()

	return &ManagerClient{
		client: pb.NewManagerClient(conn),
	}, nil
}

func (c *ManagerClient) GetCommandDefinitions() ([]*pb.CommandDefinition, error) {
	commandDefinitionList, err := c.client.GetCommandDefinitions(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return commandDefinitionList.GetCommandDefinitions(), nil
}
