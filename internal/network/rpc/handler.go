package rpc

import (
	"context"

	"google.golang.org/grpc"
)

// This method will be called to handle a new method request
type serverHandlerFunc func(s *ServerStream) error

// Represent the handler generated from the proto file
type methodHandler func(
	srv interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error)

// The interceptor is passsed to the service's handler to deal with the output of the method after it was called
func serverUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// This is the actual implemented method of the service
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}
	return nil, err
}

// Create an handler for unary requests that will call the service's handler, give it the appropriate receiver method and interceptor
func serverUnaryHandler(service interface{}, handler methodHandler) serverHandlerFunc {
	return func(stream *ServerStream) error {
		ctx := grpc.NewContextWithServerTransportStream(stream.ctx, stream)

		// The service's handler will wait for a message using the passed receiveMessage function, execute the method and the
		// output of the method will be dealt by the passed interceptor function
		response, err := handler(service, ctx, stream.receiveMessage, serverUnaryInterceptor)
		if err != nil {
			// TODO Log errors
			return nil
		}

		err = stream.sendMessage(response)
		if err != nil {
			// TODO Log errors
			return nil
		}

		return nil
	}
}
