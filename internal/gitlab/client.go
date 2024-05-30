package gitlab

import (
	"context"
	"fmt"
	"os"
	"sync"

	gl "github.com/xanzy/go-gitlab"
)

var (
	lock                          = &sync.Mutex{}
	clients map[string]*gl.Client = make(map[string]*gl.Client)
)

// GetAdminClient creates a Gitlab client with the specified token if it does not yet exist. the returned client is
// a singleton client associated with the specified `host`. If the client already exists, the token is ignored.
// This singleton is needed for reading tokens from a gitlab secret reference.
func GetAdminClient(ctx context.Context, host string, token string) (client *gl.Client, err error) {
	client, ok := clients[host]
	if ok {
		return
	}

	lock.Lock()
	defer lock.Unlock()
	client, ok = clients[host]
	if ok && client != nil {
		return
	}

	if token == "" {
		token = os.Getenv("GITLAB_TOKEN")
	}

	client, err = gl.NewClient(token, gl.WithBaseURL(fmt.Sprintf("https://%s", host)))
	clients[host] = client
	return
}
