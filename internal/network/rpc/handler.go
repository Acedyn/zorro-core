package rpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// This method will be called to handle a new method request
type serverHandlerFunc func(s *Stream) error

// The interceptor is passsed to the service's handler to deal with the output of the method after it was called
func unaryRequestInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// This is the actual implemented method of the service
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}
	return nil, err
}

// Create an handler for unary requests that will call the service's handler, give it the appropriate receiver method and interceptor
func unaryRequestHandler(service interface{}, desc *grpc.MethodDesc) serverHandlerFunc {
	return func(stream *Stream) error {
		ctx := grpc.NewContextWithServerTransportStream(stream.ctx, stream)

		// The service's handler will wait for a message using the passed RecvMsg function, execute the method and the
		// output of the method will be dealt by the passed interceptor function
		response, err := desc.Handler(service, ctx, stream.RecvMsg, unaryRequestInterceptor)
		if err != nil {
			return stream.sendResponseTrailer(fmt.Errorf("an error occured when executing the unary method handler: %w", err))
		}

		err = stream.sendResponseMessage(response)
		if err != nil {
			return fmt.Errorf("could not send message for unary request handler: %w", err)
		}
		return nil
	}
}

func streamRequestHandler(service interface{}, desc *grpc.StreamDesc) serverHandlerFunc {
	return func(stream *Stream) error {
		// The service's handler will run the implemented procedure and pass the stream to it
		err := desc.Handler(service, &ServerStream{Stream: stream})

		if err != nil {
			return stream.sendResponseTrailer(fmt.Errorf("an error occured when executing the stream method handler: %w", err))
		}
		return stream.sendResponseTrailer(nil)
	}
}
