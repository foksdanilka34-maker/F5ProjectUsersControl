package service

import (
	businesspb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/business"
	identitypb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/identity"
	"google.golang.org/grpc"
)

type Clients struct {
	Identity identitypb.IdentityServiceClient
	Business businesspb.BusinessServiceClient
}

func NewClients(identityConn, businessConn *grpc.ClientConn) *Clients {
	return &Clients{
		Identity: identitypb.NewIdentityServiceClient(identityConn),
		Business: businesspb.NewBusinessServiceClient(businessConn),
	}
}


