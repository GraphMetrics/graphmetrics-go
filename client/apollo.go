package client

import (
	"context"
	"net/http"
)

const apolloClientDetailsKey = "apollo_client_details"

func ApolloExtractor(ctx context.Context) Details {
	if details, ok := ctx.Value(apolloClientDetailsKey).(Details); ok {
		return details
	}
	return Details{}
}

func ApolloMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		details := Details{
			Name:    r.Header.Get("apollographql-client-name"),
			Version: r.Header.Get("apollographql-client-version"),
		}
		ctx = context.WithValue(ctx, apolloClientDetailsKey, details)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
