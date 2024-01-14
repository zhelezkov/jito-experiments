package jito

import (
	"context"

	mev "jito-bot/pkg/jito/gen"

	"github.com/gagliardetto/solana-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const authorizationHeader = "authorization"

type authInterceptor struct {
	token string
}

func (a *authInterceptor) UnaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return invoker(metadata.AppendToOutgoingContext(ctx, authorizationHeader, a.token), method, req, reply, cc, opts...)
}

func (a *authInterceptor) StreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return streamer(metadata.AppendToOutgoingContext(ctx, authorizationHeader, a.token), desc, cc, method, opts...)
}

func NewGrpcAuthHandler(url string, authKey solana.PrivateKey) (*authInterceptor, error) {
	authConn, err := grpc.Dial(url, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	if err != nil {
		return nil, err
	}
	defer authConn.Close()

	authClient := mev.NewAuthServiceClient(authConn)

	publicKey := authKey.PublicKey()
	res, err := authClient.GenerateAuthChallenge(context.Background(), &mev.GenerateAuthChallengeRequest{Role: mev.Role_SEARCHER, Pubkey: publicKey[:]})
	if err != nil {
		return nil, err
	}

	challenge := publicKey.String() + "-" + res.Challenge

	signed, err := authKey.Sign([]byte(challenge))
	if err != nil {
		return nil, err
	}

	tokens, err := authClient.GenerateAuthTokens(context.Background(), &mev.GenerateAuthTokensRequest{Challenge: challenge, ClientPubkey: publicKey[:], SignedChallenge: signed[:]})
	if err != nil {
		return nil, err
	}

	return &authInterceptor{
		token: "Bearer " + tokens.AccessToken.Value,
	}, nil
}
